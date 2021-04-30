import React, {useEffect} from "react";
import randomize from "randomatic";
import {Grid, makeStyles, TextField} from "@material-ui/core";
import Button from "@material-ui/core/Button";

const useStyles = makeStyles((theme) => ({
    centerAlign: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: "5px"
    },
    logo: {
        flexGrow: 1,
    },
    button: {
        color: "white",
        fontSize: "30px",
        fontFamily: "'Press Start 2P', cursive",
        margin: "50px",
        padding: "50px",
        border: "dashed 5px white",
        height: "100px",
        borderRadius: 0,
    },
    gameFont: {
        fontFamily: "'Press Start 2P', cursive",
    }
}));

const CreateRoom = (props) => {

    const classes = useStyles();

    const createParty = () => {
        const id = randomize('A0', 6, {exclude: "0oO"});
        props.history.push(`/${id}`);
    }

    const joinParty = (roomID) => {
        props.history.push(`/${roomID}`);
    }

    return (
        <Grid container className={classes.centerAlign} style={{height: "100vh"}}>
                <Button className={classes.button} onClick={createParty}>
                    CREATE PARTY
                </Button>
                <Button className={classes.button} onClick={joinParty}>
                    JOIN PARTY
                </Button>
        </Grid>
    );
};

export default CreateRoom;
