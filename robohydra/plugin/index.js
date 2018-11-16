var RoboHydraWebSocketHead = require("robohydra").heads.RoboHydraWebSocketHead;
var fs = require('fs');
exports.getBodyParts = function (conf) {
    return {
        heads: [
            new RoboHydraWebSocketHead({
                path: '/.*',
                handler: function (req, socket) {
                    var message = fs.readFileSync('robohydra/plugin/test.json');
                    socket.send(message);
                    socket.close();
                }
            })
        ]
    };
};