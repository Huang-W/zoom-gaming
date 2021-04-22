"use strict";
const jspb = require('google-protobuf');
const echo = require('./proto/echo_pb');
const input = require('./proto/input_pb');

var DCLabel = {
  /**
   * @enum {number}
   */
  Label: {
    UNDEFINED_LABEL: 0,
    ECHO: 1,
    GAME_INPUT: 2
  },
  /**
   * @param {number} label
   * @return {string}
   */
  String: (label) => {
     if (label == DCLabel.Label.ECHO)
       return "Echo";
     else if (label == DCLabel.Label.GAME_INPUT)
       return "GameInput";
     else
       throw new Error("Invalid label: " + label);
  },
  /**
   * @param {number} label
   * @return {number}
   */
  Id: (label) => {
    if (label == DCLabel.Label.GAME_INPUT)
      return 0;
    else if (label == DCLabel.Label.ECHO)
      return 1;
    else
      throw new Error("Invalid label: " + label);
  },
  /**
   * @param {number} label
   * @return {!proto.Message}
   */
  PbMessageType: (label) => {
    if (label == DCLabel.Label.ECHO)
      return echo.Echo;
    else if (label == DCLabel.Label.GAME_INPUT)
      return input.InputEvent;
    else
      throw new Error("Invalid label: " + label);
  }
};

Object.freeze(DCLabel);

export { DCLabel };
