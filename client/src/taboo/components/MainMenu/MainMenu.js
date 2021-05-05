import React from 'react'
import PropTypes from 'prop-types'
import { TabooCard } from '../shared/TabooCard'
import TextLink from '../shared/TextLink'


const MainMenu = ({ match }) => {
	const { url } = match

	const cardInfo = {
		tabooWord: 'Menu',
		list: [
			<TextLink to={`${url}/create`} text={'Create New Game'} />,
			<TextLink to={`${url}/join`} text={'Join Game'} />,
			<TextLink to={`${url}/rules`} text={'How to Play'} />,
			<TextLink to={`${url}/submit`} text={'Submit a Card'} />,
		],
	}
	return <TabooCard {...cardInfo} />
}

MainMenu.propTypes = {
	match: PropTypes.object.isRequired,
}

export default MainMenu
