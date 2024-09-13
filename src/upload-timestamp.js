const http = require('http');
const { S3Client, PutObjectCommand } = require('@aws-sdk/client-s3');

// Environment variables or defaults
const port = process.env.PORT || 8080;
const bucketName = 'gen1-test-app-write-timestamp';
const fileName = 'timestamp.txt';

// Variables to keep track of success and error counts
let successCount = 0;
let errorCount = 0;
let lastError = null; // Store the last error encountered

// Configure AWS SDK (no credentials needed when using IAM role)
const s3Client = new S3Client();

// Function to upload timestamp to S3
const uploadTimestamp = async () => {
  const timestamp = new Date().toISOString();
  
  const params = {
      Bucket: bucketName,
      Key: fileName,
      Body: `Timestamp: ${timestamp}`,
      ContentType: 'text/plain',
  };

  try {
      const command = new PutObjectCommand(params);
      await s3Client.send(command);
      successCount++;  // Increment success count on successful upload
      console.log(`Timestamp uploaded successfully: ${timestamp}`);
  } catch (err) {
      errorCount++;  // Increment error count on failure
      lastError = err;  // Store the last error encountered
      console.error('Error uploading timestamp:', err);
  }
};

// Set interval for uploading every 30 seconds
setInterval(() => {
    uploadTimestamp();
}, 30000);  // 30000 milliseconds = 30 seconds

// HTTP Server for Health Check and Stats
const requestHandler = (req, res) => {
  if (req.url === '/health') {
    // Health Check Route
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('OK');
  } else if (req.url === '/stats') {
    // Stats Route
    res.writeHead(200, { 'Content-Type': 'application/json' });
    const stats = {
      successCount,
      errorCount,
      lastError: lastError ? lastError.message : 'None',
    };
    res.end(JSON.stringify(stats));  // Return stats as a JSON object
  } else {
    // Default Route
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end('Hello, World!');
  }
};

const server = http.createServer(requestHandler);

server.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
