## How to Run

1. Open a terminal, navigate to the directory with `main.go`, and run:
   'go run main.go'

2. Open a new tab and run the following: 
   curl -X POST http://localhost:8080/receipts/process \
     -H "Content-Type: application/json" \
     -d @/path/to/examples/FILENAME.json

3. Use the output value to then execute the following:
   curl http://localhost:8080/receipts/{unique_id}/points



To check if a json is valid:
jq . /path/to/examples/FILENAME.json

