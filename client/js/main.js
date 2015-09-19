/* -*- js2-basic-offset: 4 -*- */
var husl = require('husl');

var codex = document.getElementById('codex');

var visCan = document.getElementById('art');
var visCtx = visCan.getContext('2d');

var hidCan = document.getElementById('hidden');
var hidCtx = hidCan.getContext('2d');

var scale = 4;

var state = {
    world: [],
    plants: {}
};

var xViewport = 0;

function worldWidth() {
    if (worldHeight() == 0) return 0;
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

// Render into the hidden canvas
function render() {
    hidCtx.fillStyle = colorMap[0];
    hidCtx.fillRect(0, 0, hidCan.width, hidCan.height);

    for (var y = 0; y < state.world.length; y++) {
        for (var x = 0; x < state.world[y].length; x++) {
            drawTile(x, y, state.world[y][x]);
        }
    }

    display();
}

function display() {
    visCtx.fillStyle = colorMap[0];
    visCtx.fillRect(0, 0, visCan.width, visCan.height);

    visCtx.drawImage(hidCan, xViewport, 0);
    if (xViewport > 0) {
        visCtx.drawImage(hidCan, xViewport - worldWidth() * scale, 0);
    } else {
        visCtx.drawImage(hidCan, xViewport + worldWidth() * scale, 0);
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

function drawTile(x, y, tile) {
    if (tile['tileType'] == 0) {
        return;
    }

    // XXX Do we want to do this without scaling, and scale when we copy to visctx?
    hidCtx.fillStyle = colorMap[tile['tileType']];
    hidCtx.fillRect(x * scale, y * scale, scale, scale);
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

function onResize() {
    hidCan.width = worldWidth() * scale;
    hidCan.height = worldHeight() * scale;

    visCan.width = window.innerWidth;
    visCan.height = worldHeight() * scale;

    runningResize = false;
    display();
}

var runningResize = false;
window.addEventListener("resize", function(e) {
    if (runningResize) { return; }
    runningResize = true;
    requestAnimationFrame(onResize);
});

onResize();

var dir = 1;
var lastX = -1;
visCan.addEventListener('mousedown', function(e) {
    lastX = e.clientX;
    dir = 0;
});

visCan.addEventListener('mousemove', function(e) {
    if (lastX == -1) { return; }
    xViewport -= lastX - e.clientX;

    if (lastX - e.clientX < 0) {
        dir = 1;
    } else if (lastX - e.clientX > 0) {
        dir = -1;
    }

    xViewport = xViewport % (worldWidth() * scale);
    lastX = e.clientX;
    display();
});

window.setInterval(function() {
    if (lastX == -1) {
        requestAnimationFrame(function() {
            if (worldWidth() == 0) return;
            xViewport += dir;
            xViewport = xViewport % (worldWidth() * scale);
            display();
        });
    }
}, 100);

visCan.addEventListener('mouseup', function(e) {
    lastX = -1;
});

