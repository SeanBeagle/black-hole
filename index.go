package main

var index = `
<script>
  const socket = new WebSocket('ws://localhost:8081/get');
  socket.addEventListener('open', function (event) {
	socket.send('Hello Server!');
  });
  
  socket.addEventListener('message', function (event) {
	console.log('Message from server ', event.data);
});
</script>
`
