document.addEventListener('DOMContentLoaded', function () {
    $("#canvas").click(function (e) {
        if (!isMarkerMode()) {
            return;
        }
        getPosition(e);
    });

    $('#canvas').mousedown(function (event) {
        switch (event.which) {
            case 3:
                console.log("right mouse pressed!")

                // boost/ increases the spheros speed for a few seconds
                if (boosts <= 0) {
                    //no boosts available
                    return
                }
                boosts--
                document.getElementById("boostsDisplay").innerHTML = boosts
                var xhttp = new XMLHttpRequest();
                xhttp.open("GET", "boost/", true);
                var obj = document.getElementById("boostUseMarker")
                obj.style.marginLeft = spheroPos.X * scaleFactor + 'px';
                obj.style.marginTop = spheroPos.Y * scaleFactor+ 'px';
                obj.style.borderColor = "rgba(43, 92, 190, 1)"
                obj.style.borderWidth = "10px"

                setTimeout(function name(params) {
                    obj.style.borderColor = "rgba(43, 92, 190, 0)"
                    obj.style.borderWidth = "0px"
                }, 250);

                xhttp.send();
                setTimeout(function () { trailColor = "#2020FF" }, 200);
                setTimeout(function () { trailColor = "#FF1010" }, 1500);
                break;
        }
    });

    $("#markerToggle").click(function (e) {
        //remove markers
        if (lastMarker != null) {
            lastMarker.remove()
        }

        for (i = 0; i < pointQueue.length; i++) {
            pointQueue[i].Marker.remove()
        }

        pointQueue = []
    });

    updateScale()

    document.getElementById("canvas").onmousemove = function (e) {
        if (isMarkerMode()) {
            return;
        }
        var mousecoords = getMousePos(e);
        //console.log(mousecoords)
        //drawCoordinates(mousecoords.x, mousecoords.y)
        sendTarget(mousecoords.x, mousecoords.y)
    };
    document.getElementById("connectionDisplay").style.color = "#FF0000"
    document.getElementById("connectionDisplay").innerText = "not connected"
    connected = false
    document.getElementById("startButton").style.opacity = "block"


    setInterval(function () {

        var xhttp = new XMLHttpRequest();
        xhttp.open("GET", "api/position/", true);
        xhttp.timeout = 60
        xhttp.ontimeout = function () {
            timeOuts++
            if (timeOuts > 3) {

                console.log("timeout detected!")
                document.getElementById("connectionDisplay").style.color = "#FF0000"
                document.getElementById("connectionDisplay").innerText = "not connected"
                document.getElementById("startButton").classList.remove("hide")
                document.getElementById("homeButton").classList.add("hide")
                document.getElementById("resetButton").classList.add("hide")
                connected = false
            }
        }

        xhttp.onload = function (e) {

            if (xhttp.responseText == "invalid") {
                document.getElementById("connectionDisplay").style.color = "#FF0000"
                document.getElementById("connectionDisplay").innerText = "not connected"
                document.getElementById("startButton").classList.remove("hide")
                document.getElementById("homeButton").classList.add("hide")
                document.getElementById("resetButton").classList.add("hide")
                connected = false
                return
            }


            document.getElementById("connectionDisplay").style.color = "#00FF00"
            document.getElementById("connectionDisplay").innerText = "connected"
            document.getElementById("startButton").classList.add("hide")
            document.getElementById("homeButton").classList.remove("hide")
            document.getElementById("resetButton").classList.remove("hide")
            connected = true
            timeOuts = 0

            var x = xhttp.responseText.split(':')[0] * scaleFactor; //x pos
            var y = xhttp.responseText.split(':')[1] * scaleFactor; //y pos
            var h = 180 - xhttp.responseText.split(':')[2];         //heading
            var c = xhttp.responseText.split(':')[3];         //collision 1=true

            positionUpdate(x / scaleFactor, y / scaleFactor, h)

            if (c == "1") {
                console.log("got collision message!")
                drawCollision(x + 632, y + 332)
            }

            //check if point from queue is done
            if (pointQueue.length > 0) {
                dist = distance(pointQueue[0], { X: x / scaleFactor, Y: y / scaleFactor })
                //console.log("distance: " + dist + " p1: " + pointQueue[0].X + ":" + pointQueue[0].Y + " p2: " + x / scaleFactor + ":" + y / scaleFactor)
                if (dist < 4) {
                    console.log("point done, deleting!")
                    var p = pointQueue.shift()
                    p.Marker.remove()
                }
            }

            var sphero = document.getElementById("sphero");
            sphero.style.webkitTransform = "rotate(" + h + "deg)";
            sphero.style.marginLeft = x + 'px';
            sphero.style.marginTop = y + 'px';
            drawCoordinates(x + 632, y + 332)
            document.getElementById("position").innerHTML = x / scaleFactor + ":" + y / scaleFactor + " | " + targetCoords

        }
        xhttp.send();
    }, 70);
}, false);

function getMousePos(e) {
    var hx = (document.documentElement.clientWidth) / 2 + 32
    var hy = (document.documentElement.clientHeight) / 2 + 32

    return { x: e.clientX - hx, y: e.clientY - hy };
}

function distance(p1, p2) {
    return Math.sqrt(Math.pow(p1.X - p2.X, 2) + Math.pow(p1.Y - p2.Y, 2))
}

function isMarkerMode() {
    return document.getElementById("markerToggle").checked;
}

function isQueueMode() {
    return isMarkerMode() && document.getElementById("queueToggle").checked;
}

var scaleFactor = 2
var lastCoordsX = 600
var lastCoordsY = 350
var targetCoords = "0:0"
var lastMarker
var pointQueue = [];
var timeOuts = 0
var trailColor = "#FF1010"
var connected = false

function getPosition(event) {
    var rect = canvas.getBoundingClientRect();
    var x = event.clientX - rect.left - rect.width / 2 - 30;
    var y = event.clientY - rect.top - rect.height / 2 + 15;

    //draw marker
    if (lastMarker != null && !isQueueMode()) {
        lastMarker.remove()
    }

    var m = document.createElement("img");
    m.src = "https://png.pngtree.com/svg/20170607/location_559448.png";
    m.setAttribute("width", "64");
    m.setAttribute("height", "64");
    m.setAttribute("class", "sphero")
    m.style.marginLeft = x + 'px';
    m.style.marginTop = y - 25 + 'px';
    document.getElementById("markers").appendChild(m);
    lastMarker = m

    pointQueue.push({ X: Math.round(x / scaleFactor), Y: Math.round(y / scaleFactor), Marker: m })

    //send request to server
    sendTarget(x, y)

}

function drawCollision(x, y) {
    var ctx = document.getElementById("canvas").getContext("2d");

    ctx.strokeStyle = "#FF1010"
    ctx.beginPath();
    ctx.arc(x, y, 3, 0, Math.PI * 2, true);
    ctx.fill();
}

function sendTarget(x, y) {
    addText = ",0"
    if (isQueueMode()) {
        addText = ",1"
    }
    targetCoords = Math.round(x / scaleFactor) + ":" + Math.round(y / scaleFactor)
    var xhttp = new XMLHttpRequest();
    xhttp.open("GET", "event/?" + Math.round(x / scaleFactor) + "," + Math.round(y / scaleFactor) + addText, false);
    xhttp.send();
}

function updateScale() {
    scale = document.getElementById("mapSize").value
    //100cm: width=1200; height=700 -> 1200 / x = 100 -> 1200 / 100 = x
    scaleFactor = 1200 / scale
    totalTime = 70 + scale / 6
}

function clearCanvas() {
    var canvas = document.getElementById("canvas");
    var ctx = canvas.getContext("2d");
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    if (lastMarker != null) {
        lastMarker.remove()
    }

    for (i = 0; i < pointQueue.length; i++) {
        pointQueue[i].Marker.remove()
    }

    pointQueue = []
    resetGame()
}

function drawCoordinates(x, y) {
    var ctx = document.getElementById("canvas").getContext("2d");

    ctx.strokeStyle = trailColor
    ctx.beginPath();
    ctx.moveTo(lastCoordsX, lastCoordsY);
    ctx.lineTo(x, y);
    //ctx.arc(x, y, 3, 0, Math.PI * 2, true);
    //ctx.fill();
    ctx.stroke();

    lastCoordsX = x;
    lastCoordsY = y;
}