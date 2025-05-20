# Deploying to Render.com

This document outlines the steps to deploy the Name Generator Web Server to [Render.com](https://render.com).

## Prerequisites

1. A Render.com account
2. Your project in a Git repository (GitHub, GitLab, or Bitbucket)

## Deployment Steps

### Method 1: Automatic Deployment from Git

1. **Connect your Git repository to Render**
   - Log in to your Render dashboard
   - Click "New" button and select "Web Service"
   - Connect your GitHub/GitLab/Bitbucket account
   - Select the repository containing this project

2. **Configure the Web Service**
   - Render will automatically detect the Go application
   - Verify these settings are correct:
     - **Environment**: Go
     - **Build Command**: `make build`
     - **Start Command**: `./bin/server`
     - **Region**: Choose the one closest to your users
   - Click "Create Web Service"

3. **Monitor Deployment**
   - Render will build and deploy your application
   - You can view build logs to monitor progress
   - Once deployed, you can access your application at the provided Render URL

### Method 2: Using render.yaml (Blueprint)

1. **Push your project with render.yaml to Git**
   - Ensure the render.yaml file is in the root of your repository
   - Commit and push to your Git repository

2. **Create a Blueprint Instance**
   - In your Render dashboard, click "New" and select "Blueprint"
   - Connect your Git repository
   - Render will automatically detect the render.yaml file
   - Review settings and click "Apply"

3. **Monitor Deployment**
   - Render will create all resources defined in your render.yaml file
   - You can monitor the progress in the Blueprint dashboard
   - Once complete, you can access your application at the provided URL

## Environment Variables

The application uses the following environment variables:

- `PORT`: Port on which the server will listen (set by Render automatically)
- `GO_ENV`: Environment setting (e.g., "production", "development")

## Custom Domains

To use a custom domain:

1. Go to your Web Service in the Render dashboard
2. Click on "Settings" and scroll to "Custom Domain"
3. Click "Add Custom Domain" and follow the instructions

## Scaling

To scale your application:

1. Go to your Web Service in the Render dashboard
2. Click on "Settings"
3. Under "Instance Type", select the appropriate instance size
4. For vertical scaling, choose a larger instance type
5. For horizontal scaling, increase the number of instances

## Monitoring

Render provides built-in monitoring:

1. Go to your Web Service in the Render dashboard
2. Click on "Metrics" to view:
   - CPU usage
   - Memory usage
   - Request count
   - Response time

## Logs

To view application logs:

1. Go to your Web Service in the Render dashboard
2. Click on "Logs"
3. You can filter logs and configure log drains for long-term storage

## Troubleshooting

If you encounter deployment issues:

1. **Build Failures**
   - Check your build logs for errors
   - Ensure your dependencies are properly configured
   - Verify your Makefile is working correctly

2. **Runtime Failures**
   - Check your application logs for errors
   - Verify environment variables are correctly set
   - Ensure your application is listening on the correct port

3. **Performance Issues**
   - Consider upgrading to a larger instance type
   - Use the metrics dashboard to identify bottlenecks
   - Implement caching and other performance optimizations

## Support

If you need help with Render-specific issues:

- Check the [Render documentation](https://render.com/docs)
- Contact Render support through their dashboard
- Visit the [Render Community Forum](https://community.render.com/)
