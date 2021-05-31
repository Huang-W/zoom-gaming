import React, {Suspense, lazy, useEffect} from 'react'
import {BrowserRouter as Router, Route, Switch, Redirect, useLocation} from 'react-router-dom'
import Container from './components/Container'
import Header from './components/Header'
import LayeredCards from './components/LayeredCards'
import LoadingSpinner from './components/shared/LoadingSpinner'
import { createGreetingMsg } from './utils/helpers'
import MainMenu from "./components/MainMenu";
import CreateGame from "./components/CreateGame";
import JoinGame from "./components/JoinGame";
import Rules from "./components/Rules";
import Submit from "./components/Submit";
//Code splitting routes
const Home = lazy(() => import('./pages/Home'))
const Waiting = lazy(() => import('./pages/Waiting'))
const PlayGame = lazy(() => import('./pages/PlayGame'))
const EndGame = lazy(() => import('./pages/EndGame'))
const NotFound = lazy(() => import('./pages/NotFound'))

const App = () => {
	const path = useLocation().pathname;
	window.localStorage.setItem("path", path);
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
						<Route exact path={`${path}/`}>
							<Redirect to={`${path}/home`}/>
						</Route>
						<Route path={`${path}/home`} component={Home} />
						<Route path={`${path}/waiting/:gamecode`} component={Waiting} />
						<Route path={`${path}/play/:gamecode`} component={PlayGame} />
						<Route path={`${path}/end/:gamecode`} component={EndGame} />
						<Route component={NotFound} />
					</Switch>
				</Suspense>
				<LayeredCards />
			</Container>
		</Router>
	)
}

export default App
