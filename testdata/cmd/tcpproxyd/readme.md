# What this is?

This project uses gopherjs to create a websocket client that connects to a TCP server and sends the string "Hi".

# How it works

You must have gopherjs and wfsd installed

	go install github.com/gopherjs/gopherjs
	go install github.com/andrebq/wfsd

After installing you can get the dependencies by using go get
	
	go get .

And to generate the JavaScript file type

	gopherjs build -o app.go.js main.go

# Creating a simple tcp server

This program needs a TCP server to connect, you can get one with netcat

	netcat -l 4000

Now, start wfsd on this directory, open your browser and type tcp://localhost:4000 on the inbox. When you click in the "Say hi" button, check the terminal running netcat and you should se the string "Hi" displayed.

Do not use this feature in production, since the current wfsd proxy allows the user to inform any TCP address.

# Why `go build` do not works

This project should be compiled using gopherjs and executed in a browser. Calling go build generate a valid binary but without a underlying browser almost everything will panic.

Check http://gopherjs.org for more information.
