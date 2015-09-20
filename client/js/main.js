/* -*- js2-basic-offset: 4 -*- */
var husl = require('husl');

var codex = document.getElementById('codex');

var visCan = document.getElementById('art');
var visCtx = visCan.getContext('2d');

var hidCan = document.getElementById('hidden');
var hidCtx = hidCan.getContext('2d');

var state = {
    world: [],
    plants: {}
};

var spores = [];
var selected = "";

var xViewport = 0;

function worldWidth() {
    if (worldHeight() == 0) return 0;
    return state.world[0].length;
}

function worldHeight() {
    return state.world.length;
}

var colorMap = [
    husl.toHex(30, 50, 50), //dirt
    husl.toHex(210, 50, 50), //sky
    husl.toHex(160, 70, 70), //plant default
    husl.toHex(344, 70, 70) //spore
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

    // Draw the spores
    for (var s = 0; s < spores.length; s++) {
        drawTile(spores[s]['location']['x'], spores[s]['location']['y'], {
            tileType: 3
        });
    }

    display();
}

function display() {
    visCtx.fillStyle = colorMap[0];
    visCtx.fillRect(0, 0, visCan.width, visCan.height);

    visCtx.drawImage(hidCan, xViewport, 0);
    visCtx.drawImage(hidCan, xViewport - worldWidth() * scale, 0);
    visCtx.drawImage(hidCan, xViewport + worldWidth() * scale, 0);
    visCtx.drawImage(hidCan, xViewport + worldWidth() * 2 * scale, 0);
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

    spores = delta['spores'];
}

function drawTile(x, y, tile) {
    if (tile['tileType'] == 0) {
        return;
    }

    if (tile['tileType'] == 2) {
        hidCtx.fillStyle = husl.toHex(state['plants'][tile['plant']['plantId']]['color'], 70, 70);
        if (tile['plant']['plantId'] === selected) {
            hidCtx.fillStyle = '#fff';
        }
    } else {
        hidCtx.fillStyle = colorMap[tile['tileType']];
    }

    // XXX Do we want to do this without scaling, and scale when we copy to visctx?
    hidCtx.fillRect(x * scale, y * scale, scale, scale);
}

var ws;

if (inEditMode) {
    ws = new WebSocket("ws://localhost:4444/local");
} else {
    ws = new WebSocket("ws://localhost:4444/global");
}

ws.onmessage = function(evt) {
    console.log("EVENT!");
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
        var json = JSON.parse(reader.result);

        if (firstFrame) {
            state = json;
            onResize(); // We may have changed scale, so pretend we resized
            render(); // Render the screen
            firstFrame = false;
            updateCodex();

        } else {
            applyDelta(json);
            updateCodex();

            render();
        }
    });

    reader.readAsText(evt.data);
};

ws.onopen = function() {
    console.log("Connection established");
};

ws.onclose = function() {
    document.getElementById('logo').innerHTML = "disconnected from server...<small style='font-size: 50%; vertical-align: top;'>(sorry)</small>";
};

window.onbeforeunload = function() {
    ws.onclose = function() {}; // disable onclose handler first
    ws.close();
};

function onResize() {
    visCtx = visCan.getContext('2d');
    hidCtx = hidCan.getContext('2d');

    if (hidCan.width != worldWidth() * scale) {
        hidCan.width = worldWidth() * scale;
        hidCan.height = worldHeight() * scale;
    }

    visCan.width = window.innerWidth;
    visCan.height = worldHeight() * scale;

    runningResize = false;
    display();
}

var runningResize = false;
window.addEventListener("resize", function(e) {
    if (runningResize) {
        return;
    }
    runningResize = true;
    requestAnimationFrame(onResize);
});

var dir = 1;
var lastX = -1;
visCan.addEventListener('mousedown', function(e) {
    lastX = e.clientX;
    dir = 0;
});

visCan.addEventListener('mousemove', function(e) {
    if (lastX == -1) {
        return;
    }
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

function updateCodex() {
    if (inEditMode) {
        return;
    }
    var codexString = "";

    var key = 0;
    codex.innerHTML = "";
    Object.keys(state.plants).forEach(function(key) {
        var editBtn = document.createElement('button');
        editBtn.setAttribute('class', 'edit-btn');
        editBtn.setAttribute('id', 'edit-' + key);
        editBtn.setAttribute('style', 'background-color: ' + husl.toHex(state.plants[key].color, 50, 50));
        editBtn.innerHTML = '<i class="fa fa-code-fork"></i>'; // XXX FIXME pen?

        editBtn.addEventListener("mouseover", function(event) {
            selected = key;
        }, true);

        editBtn.addEventListener("mouseout", function(event) {
            selected = "";
        }, true);

        editBtn.addEventListener("click", function(event) {
            sessionStorage.setItem('code', state.plants[key].source);
            document.location = '/edit';
        }, true);

        codex.appendChild(editBtn);
    });
}

if (inEditMode) {
    document.getElementById('test').addEventListener('click', function(e) {
        e.preventDefault();
        var v = JSON.stringify({
            kind: "+species+spawn",
            color: Math.floor(Math.random() * 360),
            code: editor.getValue(),
            ground: true
        });
        console.log(v);

        ws.send(v);
    });

    document.getElementById('publish').addEventListener('click', function(e) {
        e.preventDefault();
        var globws = new WebSocket("ws://localhost:4444/global");
        globws.onopen = function() {
            var v = JSON.stringify({
                kind: "+species+spawn",
                color: Math.floor(Math.random() * 360),
                code: editor.getValue(),
                ground: false
            });
            console.log(v);

            globws.send(v);
            globws.close();
            setTimeout(function() {
                document.location = '/';
            }, 0);
        };
    });
}
