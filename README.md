## Web app Implemantation Template

This is a template repository for implementing web applications using Go and Next.js.

### Features

- Backend: Go (Echo)
- Frontend: Next.js (React)
- Database: MySQL
- ORM: GORM
- Authentication: NextAuth(Google)

### Prerequisites

- Go 1.25
- Node.js 18+
- MySQL 8+
- pnpm 10+
- Docker
- Docker Compose

### Setup Instructions

1. Clone the repository:

   ```bash
   git clone  <repository_url>
   cd go-template-api
   ``` 
2. Set up the backend:  
3. Set up the frontend:
4. Run the application using Docker Compose:

   ```bash
   docker-compose up --build
   ```
5. Access the application at `http://localhost:3000`.
6. Stop the application:

   ```bash
   docker-compose down
   ```
### Configuration

- Update environment variables in the `.env` files for both backend and frontend as needed.
- Modify database connection settings in the backend configuration.
- Customize authentication providers in the frontend configuration.
- Adjust Docker Compose settings for your development environment.
- Refer to the documentation for more detailed configuration options.
- Feel free to contribute to this template by submitting issues or pull requests.
- Happy coding!

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details
