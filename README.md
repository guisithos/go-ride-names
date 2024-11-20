# Go Ride Names 🚴‍♂️

A fun Go application that automatically renames your Strava activities with witty jokes based on the activity type.

## Features 🌟

- Automatically detects default Strava activity names
- Assigns activity-specific jokes as new names
- Supports multiple activity types:
  - 🏃‍♂️ Running
  - 🚴‍♂️ Cycling
  - 🏊‍♂️ Swimming
  - 🚶‍♂️ Walking
  - 🏋️ Weight Training
  - 🧘‍♀️ Yoga

## Prerequisites 📋

- Go 1.16 or higher
- A Strava account
- Strava API credentials

## Setup 🔧

### 1. Get Strava API Credentials

1. Go to [Strava API Settings](https://www.strava.com/settings/api)
2. Create a new application
   - Application Name: Go Ride Names (or your preferred name)
   - Website: http://localhost
   - Authorization Callback Domain: localhost
3. After creating, you'll get:
   - Client ID
   - Client Secret

### 2. Configure Environment

1. Clone the repository:

   ```bash
   git clone https://github.com/guisithos/go-ride-names.git
   cd go-ride-names
   ```
2. Create a `.env` file in the project root:

   ```bash
   STRAVA_CLIENT_ID=<your_client_id>
   STRAVA_CLIENT_SECRET=<your_client_secret>
   ```


### 3. Authentication

1. Run the application:
   ```bash
   go run cmd/auth/main.go
   ```
2. Your browser will open to the Strava authorization page
3. Log in to Strava and authorize the application
4. After authorization, you'll receive new tokens in the terminal
5. Update your `.env` file with ALL the tokens:

   ```bash
   STRAVA_CLIENT_ID=your_client_id
   STRAVA_CLIENT_SECRET=your_client_secret
   STRAVA_ACCESS_TOKEN=your_access_token
   STRAVA_REFRESH_TOKEN=your_refresh_token
   ```


## Usage 🚀

1. Make sure your `.env` file is properly configured with all tokens
2. Run the main program:

   ```bash
   go run cmd/main.go
   ```

The program will:
- Connect to your Strava account
- Fetch your last 30 days of activities
- Detect activities with default names
- Replace them with fun, activity-specific jokes
- Show you the before and after names for each change