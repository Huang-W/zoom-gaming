import React, {useEffect} from "react";
import randomize from "randomatic";

const CreateRoom = (props) => {

    useEffect(() => {
        const id = randomize('A0', 6, {exclude: "0oO"});
        props.history.push(`/${id}`);
    })

    return (
        <></>
    );
};

export default CreateRoom;
