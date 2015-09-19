var husl = require('husl')

var codex = document.getElementById('codex');

var canvas = document.getElementById('art');
var ctx = canvas.getContext('2d');


var sx = canvas.width / 50;
var sy = canvas.height / 50;

// var colorMap = [30, 210, 160];
var colorMap = [husl.toHex(30, 50, 50),
    husl.toHex(210, 50, 50),
    husl.toHex(160, 70, 70)
]

var firstFrame = true;

function render(world) {

    for (var y = 0; y < world.length; y++) {
        for (var x = 0; x < world[y].length; x++) {
            drawTile(x, y, world[y][x]['tileType']);
        }
    }

}

function renderDelta(delta) {
    tiles = delta['tileDiff'];
    for (var i = 0; i < tiles.length; i++) {
        // tiles[i]['tile']['tileType'] = Math.round(Math.random() * 2);
        console.log(tiles[i]['tile']['tileType']);
        drawTile(tiles[i]['loc']['x'], tiles[i]['loc']['y'], tiles[i]['tile']['tileType']);

    }
}

function drawTile(x, y, type) {
    ctx.fillStyle = colorMap[type];
    ctx.fillRect(x * sx, y * sy, sx, sy);
}


var ws = new WebSocket("ws://localhost:4444/");


ws.onmessage = function(evt) {

    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
        json = JSON.parse(reader.result)
            // console.log(json);

        if (firstFrame) {

            gameheight = json['world'].length;
            gamewidth = json['world'][0].length;

            // ctx.canvas.width = window.innerWidth;
            // ctx.canvas.height = window.innerHeight;

            sx = canvas.width / gamewidth;
            sy = canvas.height / gameheight;

            render(json['world']);
            firstFrame = 0;

        } else {
            // console.log('renderingdelta')
            // console.log(json['tileDiff'])
            renderDelta(json)
        }

    });
    reader.readAsText(evt.data);

};

ws.onopen = function() {
    console.log("Connection established");
};


// ws.onclose = function() {
//     // websocket is closed.
//     console.log("Connection is closed...");
// };


window.onbeforeunload = function() {
    ws.onclose = function() {}; // disable onclose handler first
    ws.close()
};

function updateCodex(plants) {
    newHtml = ""
    for (var i = 0; i < plants.length; i++) {
        newHtml += "<div class='card'><p>" + plants[i].toString() + "<br></p></div>";
    }
    codex.innerHTML = newHtml;
}

var fakeplants = [];
fakeplants.push("test");
fakeplants.push("test2");
// fakeplants.push({firstName:"John", lastName:"Doe", age:50, eyeColor:"blue"});
updateCodex(fakeplants);
