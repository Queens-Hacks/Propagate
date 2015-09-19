var husl = require('husl')

var canvas = document.getElementById('art');
var ctx = canvas.getContext('2d');

var tilesize = 1;
var n = 0;

var sx = canvas.width/50;
var sy = canvas.height/50;

var colorMap = [30,210];

// (function animloop() {
//     requestAnimationFrame(animloop);
//     render();
//     // n++
// })();

function render(world) {

    for (var y = 0; y < world.length; y++) {
        for (var x = 0; x < world[y].length; x++) {
            // console.log(world[y][x])
            ctx.fillStyle = husl.toHex(colorMap[world[y][x]['tileType']] , 50, 50);
            ctx.fillRect(x *sx , y*sy , sx, sy);
        }
    }

    // for (var x = 0; x < canvas.width; x += tilesize) {
    //     for (var y = 0; y < canvas.height; y += tilesize) {
    //         // n = (n + 1) % 6;
    //         // ctx.fillStyle = husl.toHex(50, 0, n * 11);
    //         n += tilesize / 2
    //         ctx.fillStyle = husl.toHex(n % 360, 50, 50);

    //         ctx.fillRect(x, y, tilesize, tilesize);

    //     }
    //     // n++;
    // }


}


var ws = new WebSocket("ws://localhost:4444/");


ws.onmessage = function(evt) {


    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
        json = JSON.parse(reader.result)
            // console.log(json);
        render(json['world']);
    });
    reader.readAsText(evt.data);

};

ws.onopen = function() {
    console.log("Connection established, handle with function");
};


ws.onclose = function() {
    // websocket is closed.
    console.log("Connection is closed...");
};
