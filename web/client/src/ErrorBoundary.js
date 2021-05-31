import React from "react";
import Button from "@material-ui/core/Button";
import {Link} from "react-router-dom";

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { error: null, errorInfo: null };
  }

  static getDerivedStateFromError = error => {
    return { errorInfo: true };
  };

  componentDidCatch(error, errorInfo) {
    // Catch errors in any components below and re-render with error message
    this.setState({
      error: error,
      errorInfo: errorInfo
    })
    // You can also log error messages to an error reporting service here
  }

  render() {
    return (this.state.errorInfo) ?
      (
        <div style={{
          color: "white",
          width: "100vw",
          height: "100vh",
          textAlign: "center",
        }}>
          <h2>Something went wrong.</h2>
          <details style={{ whiteSpace: 'pre-wrap', color: "white" }}>
            {this.state.error && this.state.error.toString()}
            <br />
          </details>
          <br />
          <Link to="/" style={{textDecoration: "none"}}>
            <Button style={{
              fontSize: "20px",
              fontFamily: "'Press Start 2P', cursive",
              padding: "20px",
              border: "dashed 5px white",
              height: "50px",
              borderRadius: 0,
              color: "white",
            }}>
              HOME
            </Button>
          </Link>
        </div>
      )
      : this.props.children;
  }
}


export default ErrorBoundary