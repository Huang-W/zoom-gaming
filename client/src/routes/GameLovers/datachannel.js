"use strict";
const jspb = require('google-protobuf');
const input = require('./proto/input_pb');

var DCLabel = {
  /**
   * @enum {number}
   */
  Label: {
    UNDEFINED_LABEL: 0,
    GAME_INPUT: 1
  },
  /**
   * @param {number} label
   * @return {string}
   */
  String: (label) => {
     if (label === DCLabel.Label.GAME_INPUT)
       return "GameInput";
     else
       throw new Error("Invalid label: " + label);
  },
  /**
   * @param {number} label
   * @return {number}
   */
  Id: (label) => {
    if (label === DCLabel.Label.GAME_INPUT)
      return 0;
    else
      throw new Error("Invalid label: " + label);
  },
  /**
   * @param {number} label
   * @return {!proto.Message}
   */
  PbMessageType: (label) => {
    if (label === DCLabel.Label.GAME_INPUT)
      return input.InputEvent;
    else
      throw new Error("Invalid label: " + label);
  }
};

Object.freeze(DCLabel);

export { DCLabel };
