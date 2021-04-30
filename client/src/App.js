import React, { Suspense, lazy, useEffect } from 'react'
import { BrowserRouter as Router, Route, Switch, Redirect } from 'react-router-dom'
import CreateRoom from "./routes/CreateRoom";
import Room from "./routes/Room";
import Container from './taboo/components/Container';
import Header from './taboo/components/Header'
import LayeredCards from './taboo/components/LayeredCards'
import LoadingSpinner from './taboo/components/shared/LoadingSpinner'
import { createGreetingMsg } from './taboo/utils/helpers'
import {createMuiTheme} from "@material-ui/core";
//Code splitting routes
// const Home = lazy(() => import('./taboo/pages/Home'))
// const Waiting = lazy(() => import('./taboo/pages/Waiting'))
// const PlayGame = lazy(() => import('./taboo/pages/PlayGame'))
// const EndGame = lazy(() => import('./taboo/pages/EndGame'))
// const NotFound = lazy(() => import('./taboo/pages/NotFound'))

function App() {

  return (
  <Router>
    <Switch>
      <Route path="/" exact component={CreateRoom} />
      <Route path="/:id" component={Room} />
      {/*<Route exact path="/">*/}
      {/*  <Redirect to="/home" />*/}
      {/*</Route>*/}
      {/*<Route path="/home" component={Home} />*/}
      {/*<Route path="/waiting/:gamecode" component={Waiting} />*/}
      {/*<Route path="/play/:gamecode" component={PlayGame} />*/}
      {/*<Route path="/end/:gamecode" component={EndGame} />*/}
      {/*<Route component={NotFound} />*/}
    </Switch>
  </Router>
  );
}

export default App;
