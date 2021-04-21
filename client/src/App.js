import React, { Suspense, lazy, useEffect } from 'react'
import { BrowserRouter as Router, Route, Switch, Redirect } from 'react-router-dom'
import CreateRoom from "./routes/CreateRoom";
import Room from "./routes/Room";
import Container from './components/Container';
import Header from './components/Header'
import LayeredCards from './components/LayeredCards'
import LoadingSpinner from './components/shared/LoadingSpinner'
import { createGreetingMsg } from './utils/helpers'
//Code splitting routes
const Home = lazy(() => import('./pages/Home'))
const Waiting = lazy(() => import('./pages/Waiting'))
const PlayGame = lazy(() => import('./pages/PlayGame'))
const EndGame = lazy(() => import('./pages/EndGame'))
const NotFound = lazy(() => import('./pages/NotFound'))

function App() {
  useEffect(() => {
    if (process.env.NODE_ENV === 'production') {
      createGreetingMsg()
    }
  }, [])

  return (
  <Router>
    <Container>
      <Route component={Header} />
      <Suspense fallback={<LoadingSpinner />}>
        <Switch>
          {/*<Route path="/" exact component={CreateRoom} />*/}
          {/*<Route path="/:roomID" component={Room} />*/}
          <Route exact path="/">
            <Redirect to="/home" />
          </Route>
          <Route path="/home" component={Home} />
          <Route path="/waiting/:gamecode" component={Waiting} />
          <Route path="/play/:gamecode" component={PlayGame} />
          <Route path="/end/:gamecode" component={EndGame} />
          <Route component={NotFound} />
        </Switch>
      </Suspense>
      <LayeredCards />
    </Container>
  </Router>
  );
}

export default App;
