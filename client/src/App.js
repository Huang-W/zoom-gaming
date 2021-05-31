import React from 'react'
import { BrowserRouter as Router, Route, Switch, Redirect } from 'react-router-dom'
import CreateRoom from "./routes/CreateRoom";
import Room from "./routes/Room";

function App() {

  return (
  <Router>
    <Switch>
      <Route path="/" exact component={CreateRoom} />
      <Route path="/:roomID/:gameID" component={Room} />
    </Switch>
  </Router>
  );
}

export default App;
