import React from 'react'
import PropTypes from "prop-types"
import { Switch, Route } from 'react-router-dom'
import MainMenu from '../../components/MainMenu'
import CreateGame from '../../components/CreateGame'
import JoinGame from '../../components/JoinGame'
import Rules from '../../components/Rules'
import Submit from '../../components/Submit'

//return one of series of routes based on current home path
export default function Home({ match }) {
	const { url } = match

	return (
			<Switch>
				<Route path={url} exact component={MainMenu} />
				<Route path={`${url}/create`} component={CreateGame} />
				<Route path={`${url}/join`} component={JoinGame} />
				<Route path={`${url}/rules`} component={Rules} />
				<Route path={`${url}/submit`} component={Submit} />
			</Switch>
	)
}

Home.propTypes = {
	match : PropTypes.object.isRequired
}
