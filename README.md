## How to run

This distributed system has a hardcoded amount of nodes = 3. To run the system, you must run three separate processes with the commands:

"go run . " (This will use the default port which is 5000)

"go run . -port 5001"

"go run . -port 5002"

To request access to the critical service you simply have to press enter (or type anything) in the desired nodes terminal. 

If multiple nodes requests access at once they will be put in a queue.


## Notes about the main.go file

Each node has to act as server and client at the same time. In the `main()` function, we both setup the node as a server, listening to its own port, and also set it up as a client, connecting to the other peers as servers. Both of these are necessary for the whole architecture to work. 

## Increasing the amount of nodes in the system
In the main() method on line 160 there is a for loop. Currently it counts to 3, if this number is increased the number of nodes in the system will also increase.
If this is done you will have to run more processes, with increased ports f.eks "go run . -port 5003" for the fourth node etc.

