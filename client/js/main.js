/* -*- js2-basic-offset: 4 -*- */
var husl = require('husl');

function assert(arg, msg) {
    if (!arg) {
        console.error(new Error("arg is not true: " + msg));
    }
}

var codex = document.getElementById('codex');
var canvas = document.getElementById('art');
var ctx = canvas.getContext('2d');

var sx = 2;
var sy = 2;
var plants = [];

var state = {
    world: [],
    plants: {}
};

var xViewport = 0;

function worldWidth() {
    return state.world[0].length;
}

function worldHeight() {
    return state.world.length;
}

var colorMap = [
    husl.toHex(30, 50, 50),
    husl.toHex(210, 50, 50),
    husl.toHex(160, 70, 70)
];

var firstFrame = true;

function render() {
    // Fill with air
    ctx.fillStyle = colorMap[0];
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    for (var y = 0; y < state.world.length; y++) {
        for (var x = 0; x < state.world[y].length; x++) {
            drawTile(x, y, state.world[y][x]);
        }
    }

}

function applyDelta(delta) {
    delta.tileDiff.forEach(function(diff) {
        state.world[diff.loc.y][diff.loc.x] = diff.tile;
    });

    Object.keys(delta.newPlants).forEach(function(key) {
        state.plants[key] = delta.newPlants[key];
    });

    delta.removedPlants.forEach(function(key) {
        delete state.plants[key];
    });
}

function renderDelta(delta) {
    tiles = delta['tileDiff'];
    for (var i = 0; i < tiles.length; i++) {
        console.log(tiles[i]['tile']['tileType']);
        drawTile(tiles[i]['loc']['x'], tiles[i]['loc']['y'], tiles[i]['tile']);

    }
}

function drawTile(x, y, tile) {
    if (tile['tileType'] == 0) {
        return;
    }

    ctx.fillStyle = colorMap[tile['tileType']];
    var drawX = Math.abs((x * sx + xViewport) % (worldWidth() * sx));
    if (drawX < canvas.width) {
        ctx.fillRect(drawX, y * sy, sx, sy);
    }
}

var ws = new WebSocket("ws://localhost:4444/");

ws.onmessage = function(evt) {
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
        var json = JSON.parse(reader.result);

        if (firstFrame) {
            state = json;

            onResize(); // We may have changed scale, so pretend we resized
            render(); // Render the screen
            firstFrame = false;
        } else {
            applyDelta(json);
            render();
        }
    });

    reader.readAsText(evt.data);
};

ws.onopen = function() {
    console.log("Connection established");
};

window.onbeforeunload = function() {
    ws.onclose = function() {}; // disable onclose handler first
    ws.close();
};

/* function updateCodex(plants) {
    var newHtml = "";
    for (var i = 0; i < state.plants.length; i++) {
        newHtml += "<div class='card'><p>" + state.plants[i].toString() + "<br></p></div>";
    }
    codex.innerHTML = newHtml;
} */

function onResize() {
    canvas.width = window.innerWidth;
    canvas.height = worldHeight() * sy;
    runningResize = false;
    render();
}

var runningResize = false;
window.addEventListener("resize", function(e) {
    if (runningResize) { return; }
    runningResize = true;
    requestAnimationFrame(onResize);
});

onResize();

var lastX = -1;
art.addEventListener('mousedown', function(e) {
    console.log(e.clientX);
    lastX = e.clientX;
});

art.addEventListener('mousemove', function(e) {
    if (lastX == -1) { return; }
    console.log(e.clientX);
    xViewport -= lastX - e.clientX;
    lastX = e.clientX;
    render();
});

art.addEventListener('mouseup', function(e) {
    console.log(e.clientX);
    lastX = -1;
});

