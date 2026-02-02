# Cloud Run へのOIDCデプロイ手順書

## 置換文字列について

このドキュメント内の以下の文字列を、実際の環境に合わせて置換してください:

- `<PROJECT_ID>`: Google CloudのプロジェクトID
- `<REGION>`: デプロイ先のリージョン (例: `asia-northeast1`)
- `<SERVICE_NAME>`: Cloud Runのサービス名 (例: `go-backend-api`)
- `<GITHUB_ORG>`: GitHubの組織名またはユーザー名
- `<GITHUB_REPO>`: GitHubのリポジトリ名
- `<WORKLOAD_IDENTITY_POOL>`: Workload Identity Poolの名前 (例: `github-pool`)
- `<WORKLOAD_IDENTITY_PROVIDER>`: Workload Identity Providerの名前 (例: `github-provider`)
- `<SERVICE_ACCOUNT_NAME>`: サービスアカウント名 (例: `github-actions-deploy`)
- `<DATABASE_URL>`: データベース接続文字列 (例: `user:password@tcp(host:3306)/dbname?parseTime=true`)
- `<GCP_PROJECT_NUMBER>`: Google CloudのプロジェクトID番号

---

## 前提条件

- Google Cloudプロジェクトが作成済み
- gcloud CLIがインストール済み
- GitHubリポジトリが存在する
- Dockerfileが作成済み
- 必要な権限を持つGoogle Cloudアカウント

---

## 1. Google Cloud の初期設定

### 1.1 gcloud CLIの初期化とログイン

```bash
# Google Cloudにログイン
gcloud auth login

# プロジェクトを設定
gcloud config set project <PROJECT_ID>

# デフォルトのリージョンを設定
gcloud config set run/region <REGION>
```

### 1.2 必要なAPIの有効化

```bash
# Cloud Run API
gcloud services enable run.googleapis.com

# Artifact Registry API
gcloud services enable artifactregistry.googleapis.com

# IAM API
gcloud services enable iam.googleapis.com

# IAM Credentials API
gcloud services enable iamcredentials.googleapis.com

# Security Token Service API
gcloud services enable sts.googleapis.com

# Cloud Build API (オプション)
gcloud services enable cloudbuild.googleapis.com
```

---

## 2. Artifact Registry の設定

### 2.1 Dockerリポジトリの作成

```bash
gcloud artifacts repositories create <SERVICE_NAME> \
  --repository-format=docker \
  --location=<REGION> \
  --description="Docker repository for <SERVICE_NAME>"
```

### 2.2 認証の設定

```bash
gcloud auth configure-docker <REGION>-docker.pkg.dev
```

---

## 3. Workload Identity Federation の設定

### 3.1 Workload Identity Poolの作成

```bash
gcloud iam workload-identity-pools create <WORKLOAD_IDENTITY_POOL> \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

### 3.2 Workload Identity Providerの作成

```bash
gcloud iam workload-identity-pools providers create-oidc <WORKLOAD_IDENTITY_PROVIDER> \
  --location="global" \
  --workload-identity-pool="<WORKLOAD_IDENTITY_POOL>" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner=='<GITHUB_ORG>'" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

### 3.3 サービスアカウントの作成

```bash
gcloud iam service-accounts create <SERVICE_ACCOUNT_NAME> \
  --display-name="Service account for GitHub Actions deployment"
```

### 3.4 サービスアカウントに権限を付与

```bash
# Cloud Run管理者権限
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:<SERVICE_ACCOUNT_NAME>@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Artifact Registry書き込み権限
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:<SERVICE_ACCOUNT_NAME>@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# サービスアカウントユーザー権限
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:<SERVICE_ACCOUNT_NAME>@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

### 3.5 Workload Identity Federationの紐付け

```bash
gcloud iam service-accounts add-iam-policy-binding \
  <SERVICE_ACCOUNT_NAME>@<PROJECT_ID>.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/<GCP_PROJECT_NUMBER>/locations/global/workloadIdentityPools/<WORKLOAD_IDENTITY_POOL>/attribute.repository/<GITHUB_ORG>/<GITHUB_REPO>"
```

---

## 4. GitHub Actions ワークフローの作成

### 4.1 ワークフローファイルの設定

このリポジトリには、GitHub Actionsのワークフロー例が `.github/workflows/deploy-cloudrun.yml` として用意されています。

**使用方法:**

- ファイル内の置換文字列を実際の値に置き換える:
   - `<PROJECT_ID>`: Google CloudのプロジェクトID
   - `<REGION>`: デプロイ先のリージョン
   - `<SERVICE_NAME>`: Cloud Runのサービス名
   - `<GCP_PROJECT_NUMBER>`: プロジェクトID番号
   - `<WORKLOAD_IDENTITY_POOL>`: Workload Identity Poolの名前
   - `<WORKLOAD_IDENTITY_PROVIDER>`: Workload Identity Providerの名前
   - `<SERVICE_ACCOUNT_NAME>`: サービスアカウント名

---

## トラブルシューティング

### ログの確認

```bash
# Cloud Runのログを確認
gcloud run services logs read <SERVICE_NAME> --region=<REGION> --limit=50
```

### よくあるエラー

#### Workload Identity Federationの認証エラー

```bash
# Workload Identity Poolの確認
gcloud iam workload-identity-pools describe <WORKLOAD_IDENTITY_POOL> --location=global

# Providerの確認
gcloud iam workload-identity-pools providers describe <WORKLOAD_IDENTITY_PROVIDER> \
  --workload-identity-pool=<WORKLOAD_IDENTITY_POOL> \
  --location=global
```

#### イメージのプッシュエラー

```bash
# 認証の再設定
gcloud auth configure-docker <REGION>-docker.pkg.dev
```

---

## セキュリティのベストプラクティス

### 認証付きサービスとして設定する場合

```bash
gcloud run deploy <SERVICE_NAME> \
  --image=<REGION>-docker.pkg.dev/<PROJECT_ID>/<SERVICE_NAME>/<SERVICE_NAME>:latest \
  --no-allow-unauthenticated \
  # ... その他のオプション
```

### カスタムサービスアカウントの使用

```bash
# Cloud Run用のサービスアカウントを作成
gcloud iam service-accounts create <SERVICE_NAME>-runtime \
  --display-name="Runtime service account for <SERVICE_NAME>"

# デプロイ時に指定
gcloud run deploy <SERVICE_NAME> \
  --service-account=<SERVICE_NAME>-runtime@<PROJECT_ID>.iam.gserviceaccount.com \
  # ... その他のオプション
```

---

## 参考リンク

- [Cloud Run ドキュメント](https://cloud.google.com/run/docs)
- [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation)
- [GitHub Actions for Google Cloud](https://github.com/google-github-actions)
