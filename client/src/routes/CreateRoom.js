import React, {useEffect, useState} from "react";
import randomize from "randomatic";
import {Grid, makeStyles, TextField, Typography} from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';
import Slide from '@material-ui/core/Slide';
import broforce from "../assets/broforce.jpeg"
import spacetime from "../assets/spacetime.jpeg"
import taboo from "../assets/taboo.jpeg"
import wheels from "../assets/wheels.png"

export const useStyles = makeStyles((theme) => ({
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
    buttonGame: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        padding: "20px",
        border: "dashed 5px white",
        height: "50px",
        borderRadius: 0,
    },
    boxControls: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        padding: "20px",
        border: "dashed 5px white",
        height: "40px",
        borderRadius: 0,
    },
    buttonGameBroforce: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        margin: "20px",
        padding: "50px",
        border: "dashed 5px white",
        height: "145px",
        width: "250px",
        borderRadius: 0,
        backgroundImage: `url(${broforce})`,
        backgroundSize:"contain",
    },
    buttonGameSpacetime: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        margin: "20px",
        padding: "50px",
        border: "dashed 5px white",
        height: "145px",
        width: "250px",
        borderRadius: 0,
        backgroundImage: `url(${spacetime})`,
        backgroundSize:"contain",
    },
    buttonGameTaboo: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        margin: "20px",
        padding: "50px",
        border: "dashed 5px white",
        height: "145px",
        width: "250px",
        borderRadius: 0,
        backgroundImage: `url(${taboo})`,
        backgroundSize:"contain",
    },
    buttonGameWheels: {
        color: "white",
        fontSize: "20px",
        fontFamily: "'Press Start 2P', cursive",
        margin: "20px",
        padding: "50px",
        border: "dashed 5px white",
        height: "145px",
        width: "250px",
        borderRadius: 0,
        backgroundImage: `url(${wheels})`,
        backgroundSize:"contain",
    },
    typography: {
        fontFamily: "'Press Start 2P', cursive",
        color: "white",
        fontSize: "20px",
    },
    gameFont: {
        fontFamily: "'Press Start 2P', cursive",
    },
    mobileDialog: {
        "& .MuiPaper-root": {
            width: "700px",
            backgroundColor: "black",
            border: "solid 5px white",
            borderRadius: 0,
            fontFamily: "'Press Start 2P', cursive",
        },
    },
    textField: {
        color: "white",
        marginTop: "8px",
        marginRight: "20px",
        "& label.Mui-focused": {
            color: "white",
        },
        "& .MuiInputLabel-outlined": {
            fontSize: "10pt",
            zIndex: 1,
            fontFamily: "'Press Start 2P', cursive",
            color: "white",
        },
        "& .MuiInputLabel-shrink": {
            transform: "translate(14px, -6px) scale(0.9)",
            color: "white",
        },
        "& .MuiOutlinedInput-root": {
            borderRadius: 0,
            color: "white",
            fontFamily: "'Press Start 2P', cursive",
            "& fieldset": {
                borderColor: "white",
                border: "dashed 5px white",
            },
            "&:hover fieldset": {
                borderColor: "white",
                border: "dashed 5px white",
            },
        },
    },
}));

export const Transition = React.forwardRef(function Transition(props, ref) {
    return <Slide direction="up" ref={ref} {...props} />;
});

const CreateRoom = (props) => {

    const classes = useStyles();
    const [roomID, updateRoomID] = useState("");
    const [code, updateCode] = useState("")
    const createParty = (gameID) => {
        const id = randomize('A0', 6, {exclude: "0oO"});
        updateCode(`${id}/${gameID}`);
    }

    const joinParty = () => {
        props.history.push(`/${roomID}`);
    }
    const [openCreate, setOpenCreate] = React.useState(false);

    const handleClickOpenCreate = () => {
        setOpenCreate(true);
    };

    const handleCloseCreate = () => {
        setOpenCreate(false);
    };

    const [openJoin, setOpenJoin] = React.useState(false);

    const handleClickOpenJoin = () => {
        setOpenJoin(true);
    };

    const handleCloseJoin = () => {
        setOpenJoin(false);
    };

    return (
        <Grid container className={classes.centerAlign} style={{height: "100vh"}}>
                <Button className={classes.button} onClick={handleClickOpenCreate}>
                    CREATE PARTY
                </Button>
                <Dialog
                  open={openCreate}
                  TransitionComponent={Transition}
                  className={classes.mobileDialog}
                  keepMounted
                  maxWidth={"md"}
                  onClose={handleCloseCreate}
                  aria-labelledby="alert-dialog-slide-title"
                  aria-describedby="alert-dialog-slide-description"
                >
                    <DialogContent className={classes.centerAlign} style={{paddingBottom: "30px"}}>
                        <Grid container>
                            <Grid container item xs={6} className={classes.centerAlign}>
                                <Button className={classes.buttonGameSpacetime} onClick={() => createParty("SpaceTime")}/>
                            </Grid>
                            <Grid container item xs={6} className={classes.centerAlign}>
                                <Button className={classes.buttonGameBroforce} onClick={() => createParty("Broforce")}/>
                            </Grid>
                            <Grid container item xs={6} className={classes.centerAlign}>
                                <Button className={classes.buttonGameTaboo} onClick={() => createParty("Taboo")}/>
                            </Grid>
                            <Grid container item xs={6} className={classes.centerAlign}>
                                <Button className={classes.buttonGameWheels} onClick={() => createParty("WackyWheels")}/>
                            </Grid>
                            {code && <Grid container item xs={12}>
                                <Grid container item xs={12} className={classes.centerAlign}>
                                    <Button className={classes.buttonGame} onClick={() => navigator.clipboard.writeText(code)}>
                                        COPY CODE
                                    </Button>
                                </Grid>
                                <Grid container item xs={12} className={classes.centerAlign}>
                                    <Button className={classes.buttonGame} onClick={() => props.history.push(`/${code}`)}>
                                        START
                                    </Button>
                                </Grid>
                            </Grid>}
                        </Grid>
                    </DialogContent>
                </Dialog>
                <Button className={classes.button} onClick={handleClickOpenJoin}>
                    JOIN PARTY
                </Button>
                <Dialog
                  open={openJoin}
                  TransitionComponent={Transition}
                  className={classes.mobileDialog}
                  keepMounted
                  onClose={handleCloseJoin}
                  aria-labelledby="alert-dialog-slide-title"
                  aria-describedby="alert-dialog-slide-description"
                >
                    <DialogContent className={classes.centerAlign} style={{margin: "20px"}}>
                        <TextField label="Room ID" variant="outlined" className={classes.textField} value={roomID} onChange={(e) => updateRoomID(e.target.value)}/>
                        <Button className={classes.buttonGame} onClick={joinParty}>
                            JOIN
                        </Button>
                    </DialogContent>
                </Dialog>
        </Grid>
    );
};

export default CreateRoom;
