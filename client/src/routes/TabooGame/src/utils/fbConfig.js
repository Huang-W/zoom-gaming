import firebase from "firebase/app"
import "firebase/firestore"
import "firebase/auth"

// Your web app's Firebase configuration
const firebaseConfig = {
  apiKey: "AIzaSyB0ZczV5f2ueSkpwIBaomuKltIvGT9iE0w",
  authDomain: "taboo-game-afaf0.firebaseapp.com",
  projectId: "taboo-game-afaf0",
  storageBucket: "taboo-game-afaf0.appspot.com",
  messagingSenderId: "424180926413",
  appId: "1:424180926413:web:c65d6f8fa3bf4e164aa684",
  measurementId: "G-5FPD4RVVBF"
};

// apiKey: JSON.stringify(process.env.REACT_APP_DB_URL),
// authDomain: JSON.stringify(process.env.REACT_APP_AUTH_DOMAIN),
// projectId: JSON.stringify(process.env.REACT_APP_PROJECT_ID),
// storageBucket: JSON.stringify(process.env.REACT_APP_STORAGE_BUCKET),
// messagingSenderId: JSON.stringify(process.env.REACT_APP_MSG_SENDER_ID),
// appId: JSON.stringify(process.env.REACT_APP_APP_ID),
// measurementId: JSON.stringify(process.env.REACT_APP_MEASUREMENT_ID)

// Initialize Firebase
firebase.initializeApp(firebaseConfig)

//Initialize firestore
firebase.firestore()

export default firebase
