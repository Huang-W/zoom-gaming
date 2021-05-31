import React from 'react'
import { createStore, compose, applyMiddleware } from 'redux'
import thunk from 'redux-thunk'
import { Provider } from 'react-redux'

import { reduxFirestore, getFirestore, createFirestoreInstance } from 'redux-firestore'
import { ReactReduxFirebaseProvider, getFirebase } from 'react-redux-firebase'
import firebase from 'firebase/app'
import fbConfig from './utils/fbConfig'
import rootReducer from './store/reducers/rootReducer.js'
import { ThemeProvider } from 'styled-components'
import App from './App.js'
import GlobalStyle from './global-design/globalStyles'
import theme from './global-design/theme'

const store = createStore(
	rootReducer,
	compose(applyMiddleware(thunk.withExtraArgument({ getFirebase, getFirestore })), reduxFirestore(fbConfig))
)

//react-redux-firebase props
const rrfProps = {
	firebase,
	config: fbConfig,
	dispatch: store.dispatch,
	createFirestoreInstance,
	presence: 'presence',
	sessions: 'sessions',
}

const TabooGame = () => {
	return (
			<Provider store={store}>
				<ReactReduxFirebaseProvider {...rrfProps}>
					<ThemeProvider theme={theme}>
						<GlobalStyle />
						<App />
					</ThemeProvider>
				</ReactReduxFirebaseProvider>
			</Provider>
			)
};
export default TabooGame;

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
