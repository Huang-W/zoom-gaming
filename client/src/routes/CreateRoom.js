import React, {useEffect} from "react";
import { v1 as uuid } from "uuid";

const CreateRoom = (props) => {

    useEffect(() => {
        const id = uuid();
        props.history.push(`/${id}`);
    })

    return (
        <></>
    );
};

export default CreateRoom;
