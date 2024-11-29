Receipt Processor Web Service

This is a Go implementation of the Receipt Processor web service that fulfills the documented API. The application stores data in memory and calculates points based on the specified rules.

Prerequisites:
Go: Ensure Go is installed on your system (version 1.13 or later)

1. Clone the Repository
   
   `git clone https://github.com/yourusername/receipt-processor.git`
   
   `cd receipt-processor`

2. Save the Code

   Save the provided `main.go` file into the receipt-processor directory. You can copy the code from main.go.

3. Initialize the Go Module

   Initialize a new Go module:
   
   `go mod init receipt-processor`

4. Download Dependencies

   Run the following command to download the necessary dependencies:
   
   `go get github.com/google/uuid`

5. Running the Application, the server will start on port 8080:

   `go run main.go`


6. Testing the Application:

    a. Submit a Receipt for Processing
    
    Make a `POST` request to `/receipts/process` with a JSON payload. Here's an example using `curl`:
    
    `curl -X POST http://localhost:8080/receipts/process \
       -H 'Content-Type: application/json' \
       -d '{
         "retailer": "Target",
         "purchaseDate": "2022-01-01",
         "purchaseTime": "13:01",
         "items": [
           {
             "shortDescription": "Mountain Dew 12PK",
             "price": "6.49"
           },
           {
             "shortDescription": "Emils Cheese Pizza",
             "price": "12.25"
           },
           {
             "shortDescription": "Knorr Creamy Chicken",
             "price": "1.26"
           },
           {
             "shortDescription": "Doritos Nacho Cheese",
             "price": "3.35"
           },
           {
             "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
             "price": "12.00"
           }
         ],
         "total": "35.35"
       }'
    `
    
    Expected response:
    `{
      "id": "your-generated-id"
    }`


      b. Retrieve Points for a Receipt
      Use the ID from the previous response to make a `GET` request to `/receipts/{id}/points`:
      
      `curl http://localhost:8080/receipts/your-generated-id/points`
      
      Expected response:
      `
      {
        "points": 28
      }
      `


Notes:

Data Persistence: The application stores data in memory. All data will be lost when the application stops.

Thread Safety: A mutex lock is used to handle concurrent access to the in-memory data store.

Validation: The application validates the receipt data against the specified patterns and formats.


Explanation of Points Calculation:

The `calculatePoints` function implements the rules:

1. Alphanumeric Characters in Retailer Name: Counts each alphanumeric character in the retailer's name.
   
2. Round Dollar Total: Adds 50 points if the total is a whole number.

3. Multiple of 0.25 Total: Adds 25 points if the total is a multiple of 0.25.

4. Items Pair: Adds 5 points for every two items.

5. Item Description Length Multiple of 3: For items with descriptions whose trimmed length is a multiple of 3, it adds 20% of the item price (rounded up) to the points.

6. Odd Purchase Day: Adds 6 points if the purchase day is odd.

7. Purchase Time Between 2:00 PM and 4:00 PM: Adds 10 points if the purchase time is after 2:00 PM and before 4:00 PM.






