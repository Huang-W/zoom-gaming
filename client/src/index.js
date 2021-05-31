import React from 'react';
import ReactDOM from 'react-dom'
import './index.css';
import * as serviceWorker from './serviceWorker';
import ParticlesBg from "particles-bg";
import App from './App'


let config = {
  num: [4, 7],
  rps: 0.1,
  radius: [0, 40],
  life: [1.5, 3],
  v: [2, 3],
  tha: [-40, 40],
  alpha: [0.6, 0],
  scale: [.1, 0.4],
  position: "all",
  color: ["random", "#ff0000"],
  cross: "dead",
  // emitter: "follow",
  random: 15,
  g: 5,    // gravity
};

if (Math.random() > 0.85) {
  config = Object.assign(config, {
    onParticleUpdate: (ctx, particle) => {
      ctx.beginPath();
      ctx.rect(
        particle.p.x,
        particle.p.y,
        particle.radius * 2,
        particle.radius * 2
      );
      ctx.fillStyle = particle.color;
      ctx.fill();
      ctx.closePath();
    }
  });
}


ReactDOM.render(
  <React.StrictMode>
    <App />
    <ParticlesBg type="custom" config={config} bg={true} />
  </React.StrictMode>,
  document.getElementById('root')
)

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
