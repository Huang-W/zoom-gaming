import React from "react"
import PropTypes from "prop-types"
import {  StyledHeader, Title, Subheading, FocusSpan} from "./style.js"


export default function Header({ location }) {
  //size of header and whether subheading included varies based on route location
  const homeOrEndRoute = location.pathname.includes("home") || location.pathname.includes("end")
  const homeRouteExact = location.pathname === ("/home")

  return (
    <></>
  )
}

Header.propTypes = {
  location: PropTypes.object.isRequired,
}
