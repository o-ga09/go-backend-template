# Amazon ECS へのデプロイ手順書

## 置換文字列について

このドキュメント内の以下の文字列を、実際の環境に合わせて置換してください:

- `<AWS_ACCOUNT_ID>`: AWSアカウントID
- `<AWS_REGION>`: デプロイ先のリージョン (例: `ap-northeast-1`)
- `<ECR_REPOSITORY_NAME>`: ECRリポジトリ名 (例: `go-backend-api`)
- `<ECS_CLUSTER_NAME>`: ECSクラスター名 (例: `go-backend-cluster`)
- `<ECS_SERVICE_NAME>`: ECSサービス名 (例: `go-backend-service`)
- `<TASK_DEFINITION_NAME>`: タスク定義名 (例: `go-backend-task`)
- `<TASK_EXECUTION_ROLE_NAME>`: タスク実行ロール名 (例: `ecsTaskExecutionRole`)
- `<TASK_ROLE_NAME>`: タスクロール名 (例: `ecsTaskRole`)
- `<VPC_ID>`: 使用するVPCのID
- `<SUBNET_ID_1>`, `<SUBNET_ID_2>`: 使用するサブネットのID
- `<SECURITY_GROUP_ID>`: セキュリティグループID
- `<ALB_TARGET_GROUP_ARN>`: Application Load BalancerのターゲットグループARN
- `<GITHUB_ORG>`: GitHubの組織名またはユーザー名
- `<GITHUB_REPO>`: GitHubのリポジトリ名
- `<DATABASE_URL>`: データベース接続文字列
- `<CONTAINER_NAME>`: コンテナ名 (例: `go-backend-container`)
- `<LOG_GROUP_NAME>`: CloudWatch Logsのロググループ名 (例: `/ecs/go-backend`)

---

## 前提条件

- AWSアカウントが作成済み
- AWS CLIがインストール済み
- GitHubリポジトリが存在する
- Dockerfileが作成済み
- VPC、サブネット、セキュリティグループが設定済み
- 必要な権限を持つIAMユーザーまたはロール

---

## 1. AWS CLIの初期設定

### 1.1 AWS CLIの設定

```bash
# AWS CLIの設定
aws configure

# 入力項目:
# - AWS Access Key ID
# - AWS Secret Access Key
# - Default region name: <AWS_REGION>
# - Default output format: json
```

---

## 2. ECR (Elastic Container Registry) の設定

### 2.1 ECRリポジトリの作成

```bash
aws ecr create-repository \
  --repository-name <ECR_REPOSITORY_NAME> \
  --region <AWS_REGION> \
  --image-scanning-configuration scanOnPush=true \
  --encryption-configuration encryptionType=AES256
```

### 2.2 ECRへのログイン

```bash
aws ecr get-login-password --region <AWS_REGION> | \
  docker login --username AWS --password-stdin <AWS_ACCOUNT_ID>.dkr.ecr.<AWS_REGION>.amazonaws.com
```

---

## 3. IAMロールの作成

### 3.1 タスク実行ロールの作成

タスク実行ロールの信頼ポリシー (`trust-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

```bash
# タスク実行ロールの作成
aws iam create-role \
  --role-name <TASK_EXECUTION_ROLE_NAME> \
  --assume-role-policy-document file://trust-policy.json

# 必要なポリシーのアタッチ
aws iam attach-role-policy \
  --role-name <TASK_EXECUTION_ROLE_NAME> \
  --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy

# ECRアクセス権限
aws iam attach-role-policy \
  --role-name <TASK_EXECUTION_ROLE_NAME> \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly
```

### 3.2 タスクロールの作成 (オプション)

```bash
# タスクロールの作成
aws iam create-role \
  --role-name <TASK_ROLE_NAME> \
  --assume-role-policy-document file://trust-policy.json
```

### 3.3 GitHub Actions用のIAMロール作成 (OIDC)

OIDC Provider用の信頼ポリシー (`github-trust-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<AWS_ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:<GITHUB_ORG>/<GITHUB_REPO>:*"
        }
      }
    }
  ]
}
```

```bash
# GitHub Actions用のOIDC Providerを作成 (初回のみ)
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# GitHub Actions用のロール作成
aws iam create-role \
  --role-name GitHubActionsDeployRole \
  --assume-role-policy-document file://github-trust-policy.json
```

### 3.4 GitHub Actionsロールにポリシーをアタッチ

カスタムポリシー (`github-actions-policy.json`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecs:UpdateService",
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:RegisterTaskDefinition"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole"
      ],
      "Resource": [
        "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<TASK_EXECUTION_ROLE_NAME>",
        "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<TASK_ROLE_NAME>"
      ]
    }
  ]
}
```

```bash
# ポリシーの作成とアタッチ
aws iam put-role-policy \
  --role-name GitHubActionsDeployRole \
  --policy-name GitHubActionsDeployPolicy \
  --policy-document file://github-actions-policy.json
```

---

## 4. CloudWatch Logs の設定

```bash
# ロググループの作成
aws logs create-log-group \
  --log-group-name <LOG_GROUP_NAME> \
  --region <AWS_REGION>

# ログの保持期間を設定 (例: 7日間)
aws logs put-retention-policy \
  --log-group-name <LOG_GROUP_NAME> \
  --retention-in-days 7 \
  --region <AWS_REGION>
```

---

## 5. ECSクラスターの作成

```bash
# Fargateクラスターの作成
aws ecs create-cluster \
  --cluster-name <ECS_CLUSTER_NAME> \
  --region <AWS_REGION> \
  --capacity-providers FARGATE FARGATE_SPOT \
  --default-capacity-provider-strategy \
    capacityProvider=FARGATE,weight=1,base=1
```

---

## 6. タスク定義の作成

### 6.1 タスク定義ファイルの作成

`task-definition.json` を作成:

```json
{
  "family": "<TASK_DEFINITION_NAME>",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<TASK_EXECUTION_ROLE_NAME>",
  "taskRoleArn": "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<TASK_ROLE_NAME>",
  "containerDefinitions": [
    {
      "name": "<CONTAINER_NAME>",
      "image": "<AWS_ACCOUNT_ID>.dkr.ecr.<AWS_REGION>.amazonaws.com/<ECR_REPOSITORY_NAME>:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        },
        {
          "name": "PORT",
          "value": "8080"
        },
        {
          "name": "PROJECT_ID",
          "value": "<ECS_SERVICE_NAME>"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:<AWS_REGION>:<AWS_ACCOUNT_ID>:secret:database-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "<LOG_GROUP_NAME>",
          "awslogs-region": "<AWS_REGION>",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### 6.2 タスク定義の登録

```bash
aws ecs register-task-definition \
  --cli-input-json file://task-definition.json \
  --region <AWS_REGION>
```

---

## 7. ECSサービスの作成

```bash
aws ecs create-service \
  --cluster <ECS_CLUSTER_NAME> \
  --service-name <ECS_SERVICE_NAME> \
  --task-definition <TASK_DEFINITION_NAME> \
  --desired-count 2 \
  --launch-type FARGATE \
  --platform-version LATEST \
  --network-configuration "awsvpcConfiguration={subnets=[<SUBNET_ID_1>,<SUBNET_ID_2>],securityGroups=[<SECURITY_GROUP_ID>],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=<ALB_TARGET_GROUP_ARN>,containerName=<CONTAINER_NAME>,containerPort=8080" \
  --health-check-grace-period-seconds 60 \
  --deployment-configuration "maximumPercent=200,minimumHealthyPercent=100" \
  --region <AWS_REGION>
```

---

## 8. Secrets Manager でのシークレット管理 (オプション)

### 8.1 シークレットの作成

```bash
# データベース接続文字列の保存
aws secretsmanager create-secret \
  --name database-url \
  --description "Database connection string" \
  --secret-string "<DATABASE_URL>" \
  --region <AWS_REGION>
```

### 8.2 タスク実行ロールに権限追加

```bash
# Secrets Managerへのアクセス権限を追加
aws iam attach-role-policy \
  --role-name <TASK_EXECUTION_ROLE_NAME> \
  --policy-arn arn:aws:iam::aws:policy/SecretsManagerReadWrite
```

---

## 9. GitHub Actions ワークフローの作成

### 9.1 ワークフローファイルの設定

このリポジトリには、GitHub Actionsのワークフロー例が `.github/workflows/deploy-ecs.yml` として用意されています。

**使用方法:**

- ファイル内の置換文字列を実際の値に置き換える:
   - `<AWS_ACCOUNT_ID>`: AWSアカウントID
   - `<AWS_REGION>`: デプロイ先のリージョン
   - `<ECR_REPOSITORY_NAME>`: ECRリポジトリ名
   - `<ECS_CLUSTER_NAME>`: ECSクラスター名
   - `<ECS_SERVICE_NAME>`: ECSサービス名
   - `<TASK_DEFINITION_NAME>`: タスク定義名
   - `<CONTAINER_NAME>`: コンテナ名

---

## 10. Application Load Balancer (ALB) の設定

### 10.1 ターゲットグループの作成

```bash
aws elbv2 create-target-group \
  --name <ECS_SERVICE_NAME>-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id <VPC_ID> \
  --target-type ip \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3 \
  --region <AWS_REGION>
```

### 9.2 ALBの作成

```bash
aws elbv2 create-load-balancer \
  --name <ECS_SERVICE_NAME>-alb \
  --subnets <SUBNET_ID_1> <SUBNET_ID_2> \
  --security-groups <SECURITY_GROUP_ID> \
  --scheme internet-facing \
  --type application \
  --ip-address-type ipv4 \
  --region <AWS_REGION>
```

### 9.3 リスナーの作成

```bash
aws elbv2 create-listener \
  --load-balancer-arn <ALB_ARN> \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=<ALB_TARGET_GROUP_ARN> \
  --region <AWS_REGION>
```

---

## 10. オートスケーリングの設定 (オプション)

### 10.1 Application Auto Scalingのターゲット設定

```bash
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/<ECS_CLUSTER_NAME>/<ECS_SERVICE_NAME> \
  --min-capacity 2 \
  --max-capacity 10 \
  --region <AWS_REGION>
```

### 10.2 スケーリングポリシーの作成

```bash
# CPU使用率ベースのスケーリング
aws application-autoscaling put-scaling-policy \
  --policy-name cpu-scaling-policy \
  --service-namespace ecs \
  --scalable-dimension ecs:service:DesiredCount \
  --resource-id service/<ECS_CLUSTER_NAME>/<ECS_SERVICE_NAME> \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration file://scaling-policy.json \
  --region <AWS_REGION>
```

`scaling-policy.json`:

```json
{
  "TargetValue": 70.0,
  "PredefinedMetricSpecification": {
    "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
  },
  "ScaleInCooldown": 300,
  "ScaleOutCooldown": 60
}
```

---

## トラブルシューティング

### ログの確認

```bash
# CloudWatch Logsの確認
aws logs tail <LOG_GROUP_NAME> --follow --region <AWS_REGION>

# タスクの停止理由を確認
aws ecs describe-tasks \
  --cluster <ECS_CLUSTER_NAME> \
  --tasks <TASK_ARN> \
  --region <AWS_REGION>
```

### よくあるエラー

#### タスクが起動しない

```bash
# サービスイベントを確認
aws ecs describe-services \
  --cluster <ECS_CLUSTER_NAME> \
  --services <ECS_SERVICE_NAME> \
  --region <AWS_REGION> \
  --query 'services[0].events[0:10]'
```

#### イメージのプルエラー

```bash
# タスク実行ロールの権限を確認
aws iam get-role --role-name <TASK_EXECUTION_ROLE_NAME>
aws iam list-attached-role-policies --role-name <TASK_EXECUTION_ROLE_NAME>
```

---

## セキュリティのベストプラクティス

### VPCエンドポイントの使用

```bash
# ECR用のVPCエンドポイント作成
aws ec2 create-vpc-endpoint \
  --vpc-id <VPC_ID> \
  --service-name com.amazonaws.<AWS_REGION>.ecr.dkr \
  --route-table-ids <ROUTE_TABLE_ID> \
  --region <AWS_REGION>

aws ec2 create-vpc-endpoint \
  --vpc-id <VPC_ID> \
  --service-name com.amazonaws.<AWS_REGION>.ecr.api \
  --route-table-ids <ROUTE_TABLE_ID> \
  --region <AWS_REGION>
```

### セキュリティグループの最小権限設定

```bash
# インバウンドルールの例
aws ec2 authorize-security-group-ingress \
  --group-id <SECURITY_GROUP_ID> \
  --protocol tcp \
  --port 8080 \
  --source-group <ALB_SECURITY_GROUP_ID> \
  --region <AWS_REGION>
```

---

## 参考リンク

- [Amazon ECS ドキュメント](https://docs.aws.amazon.com/ecs/)
- [AWS Fargate](https://docs.aws.amazon.com/fargate/)
- [Amazon ECR](https://docs.aws.amazon.com/ecr/)
- [GitHub Actions for AWS](https://github.com/aws-actions)
- [AWS OIDC with GitHub Actions](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
