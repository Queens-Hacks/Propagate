var husl = require('husl')

var canvas = document.getElementById('art');
var ctx = canvas.getContext('2d');

var tilesize = 10;
var n = 0;



// (function animloop() {
//     requestAnimationFrame(animloop);
//     render();
//     // n++
// })();

function render() {
    for (var x = 0; x < canvas.width; x += tilesize) {
        for (var y = 0; y < canvas.height; y += tilesize) {
            // n = (n + 1) % 6;
            // ctx.fillStyle = husl.toHex(50, 0, n * 11);
            n += tilesize / 2
            ctx.fillStyle = husl.toHex(n % 360, 50, 50);

            ctx.fillRect(x, y, tilesize, tilesize);

        }
        // n++;
    }


}


var ws = new WebSocket("ws://localhost:4444/");


ws.onmessage = function(evt) {
    var received_msg = evt.data;
    console.log(received_msg.toString())
        // alert("Message is received...");
};

ws.onopen = function() {
    alert("Connection established, handle with function");
};


ws.onclose = function() {
    // websocket is closed.
    alert("Connection is closed...");
};
