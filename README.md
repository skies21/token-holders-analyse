## Token Holders Analysis
This project is designed to analyze token holders for a given token on the Solana blockchain. 
It fetches data about token holders and their transactions, processes it, and provides insights into token distribution and transaction history.
# Prerequisites
Make sure you have the following installed on your system:
* Docker
* Go
# Usage
1. Clone the repository to your local machine:
    ```
    git clone https://github.com/your-username/token-holders-analysis.git
    ```
2. Navigate to the project directory:
    ```
    cd token-holders-analysis
    ```
3. Create a .env file in the project root directory and set the following environment variables:
    ```
    HELIUS_API=your-helius-api-key
    RATE_LIMIT=100 # recommended by default solanaFM rate limit
   ```
4. Build and run Docker :
    ```
   docker-compose up --build
   ```
5. Access the web interface at http://localhost:3000 in your browser. 
6. Try passing token hash, for example, http://127.0.0.1:3000/EMcz7rjNJatWAPvG34iPgrwhcnfZdBWKJQFR1b6rCWT2