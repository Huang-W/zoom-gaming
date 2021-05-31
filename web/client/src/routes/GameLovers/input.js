"use strict";
const pb = require('./proto/input_pb');

var InputMap = new Map([
  ['ArrowUp', pb.KeyPressEvent.Key.KEY_ARROW_UP],
  ['ArrowDown', pb.KeyPressEvent.Key.KEY_ARROW_DOWN],
  ['ArrowLeft', pb.KeyPressEvent.Key.KEY_ARROW_LEFT],
  ['ArrowRight', pb.KeyPressEvent.Key.KEY_ARROW_RIGHT],
  ['Space', pb.KeyPressEvent.Key.KEY_SPACE],
  ['KeyA', pb.KeyPressEvent.Key.KEY_KEY_A],
  ['KeyS', pb.KeyPressEvent.Key.KEY_KEY_S],
  ['KeyD', pb.KeyPressEvent.Key.KEY_KEY_D]
]);

Object.freeze(InputMap);

export { InputMap };
