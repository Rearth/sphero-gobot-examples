document.addEventListener('DOMContentLoaded', function () {
    setTarget()
    createBoost()

    if (score == null) {
        score = document.getElementById("pointsDisplay")
    }

    setInterval(frameUpdate, 100)
})

var curTarget = null
var spheroPos = {X: 0, Y: 0}
var score = null
var points = 0
var progress = 0
var totalTime = 100.0
var boosts = 0
var boostPos = null
var boostChance = 0.014
var misses = 0
var lost = false

function boostCollected() {
    boosts++
    document.getElementById("boostsDisplay").innerHTML = boosts
    displayReachAnim(boostPos.X, boostPos.Y)
    boostPos = null
    document.getElementById("boostIcon").style.display = "none"
}

function createBoost() {
    width = document.getElementById("mapSize").value
    height = 700 / 1200 * width

    x = Math.floor(Math.random() * width * 0.8) - width / 2
    y = Math.floor(Math.random() * height * 0.8) - height / 2 + 10
    console.log("setting boost at: x=" + x + " y=" + y)
    boostPos = {X: x, Y: y}

    document.getElementById("boostIcon").style.display = "block"

    drawPoint(boostPos.X, boostPos.Y, "transparent", 0, "boostIcon")
}

function resetGame() {
    points = 0
    progress = 0
    totalTime = 100
    boosts = 0
    misses = 0
    boosts = 0
    lost = false
    setTarget()
    createBoost()
    document.getElementById("boostsDisplay").innerHTML = boosts
    document.getElementById("missesDisplay").innerHTML = misses
    document.getElementById("state").innerHTML = ""
    score.innerHTML = points

    document.getElementById("separator").style.display = "block"
    
}

function frameUpdate() {
    if (!connected || lost) {
        return
    }
    progress++
    if (progress <= totalTime) {
        i = 220 * (progress / totalTime)
        color="rgb(" + i + ", " + i + ", 255)"
        
        drawPoint(curTarget.X, curTarget.Y, color, progress / totalTime)
    } else {
        misses++
        document.getElementById("missesDisplay").innerHTML = misses
        //point missed
        if (misses >= 3) {
            //game lost, too many consecutive misses
            gameLost()
            return
        }
        color = "white"
        drawPoint(curTarget.X, curTarget.Y, color)
        setTarget()
        totalTime++
        progress = 0


    }

    if (boostPos == null && Math.random() < boostChance) {
        createBoost()
    }
}

function gameLost() {
    console.log("game lost!")
    lost = true
    document.getElementById("state").innerHTML = "Game Lost! Points: " + points
    document.getElementById("separator").style.display = "none"
}

function setTarget() {

    width = document.getElementById("mapSize").value
    height = 700 / 1200 * width

    x = Math.floor(Math.random() * width * 0.8) - width / 2
    y = Math.floor(Math.random() * height * 0.8) - height / 2 + 10
    console.log("setting target at: x=" + x + " y=" + y)
    // var m = document.createElement("img");
    // m.src = "images/star-flat.png";
    // m.setAttribute("width", "32");
    // m.setAttribute("height", "32");
    // m.setAttribute("class", "sphero")
    // m.style.marginLeft = x * scaleFactor * 0.95 - 16 + 'px';
    // m.style.marginTop = y * scaleFactor * 0.9 + 16 +'px';
    // document.getElementById("markers").appendChild(m);

    //drawPoint(x, y, "#000000")
    curTarget = { X: x, Y: y }

}

function targetReached() {

    console.log("target reached")
    misses = 0
    totalTime *= 0.94
    points++
    document.getElementById("missesDisplay").innerHTML = misses
    score.innerHTML = points

    displayReachAnim(curTarget.X, curTarget.Y)

    drawPoint(curTarget.X, curTarget.Y, "white")
    //curTarget.marker.remove()
    setTarget()

    progress = 0
}

function displayReachAnim(x, y) {
    var obj = document.getElementById("reachedMarker")
    obj.style.marginLeft = x * scaleFactor + 'px';
    obj.style.marginTop = y * scaleFactor + 'px';
    obj.style.borderColor = "rgba(43, 190, 92, 1)"
    obj.style.borderWidth = "10px"

    setTimeout(function name(params) {
        obj.style.borderColor = "rgba(43, 190, 92, 0)"
        obj.style.borderWidth = "0px"
    }, 250);
}

function drawPoint(x, y, color, progress = 0.95, id="targetMarker") {
    obj = document.getElementById(id);
    obj.style.marginLeft = x * scaleFactor + 10 + 'px';
    obj.style.marginTop = y * scaleFactor + 20 + 'px';
    obj.style.background = color
    animate($('.' + id), 0, progress);

    //obj.style.borderColor = color

    // degree = degree * Math.PI / 180

    // var ctx = document.getElementById("canvas").getContext("2d");
    // xDraw = x * scaleFactor + 632
    // yDraw = y * scaleFactor + 332
    // ctx.beginPath();
    // ctx.fillStyle = color
    // ctx.arc(xDraw, yDraw, size, degree, 0, true);
    // ctx.fill()
    // ctx.closePath()
}

function positionUpdate(x, y, h) {
    mapSize = document.getElementById("mapSize").value
    if (lost || !connected) {
        return
    }
    //200 * x = 6

    reqDist = mapSize * 0.03
    spheroPos = {X: x, Y: y}
    //console.log("game update: X=" + x + " y=" + y + " target=" + curTarget.X + ":" + curTarget.Y)
    if (distance(curTarget, spheroPos) < reqDist) {
        targetReached()
    }

    if (boostPos != null && distance(boostPos, spheroPos) < reqDist) {
        boostCollected()
    }

}
