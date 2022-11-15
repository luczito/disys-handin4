## How to run

This distributed system has a hardcoded amount of nodes = 3. To run the system, you must run three separate processes with the commands:

"go run . " (This will use the default port which is 5000)

"go run . -port 5001"

"go run . -port 5002"

To request access to the critical service you simply have to press enter (or type anything) in the desired nodes terminal. 

If multiple nodes requests access at once they will be put in a queue.