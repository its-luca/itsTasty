import Typography from "@mui/material/Typography";
import {
    Dialog, DialogActions, DialogTitle,
    List,
    ListItem, ListItemAvatar, ListItemText,
    Paper,
    SxProps,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableRow,
    Theme, TextField, AlertTitle, Alert, Link
} from "@mui/material";
import Container from "@mui/material/Container";
import {useAuthContext} from "../AuthContext";
import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query";
import {ApiError, DefaultService, MergedDishManagementData, MergedDishUpdateReq} from "../services/userAPI";
import Avatar from "@mui/material/Avatar";
import IconButton from "@mui/material/IconButton";
import DeleteIcon from '@mui/icons-material/Delete';
import Tooltip from "@mui/material/Tooltip";
import {ChangeEvent, useState} from "react";
import Button from "@mui/material/Button";
import {useNavigate} from "react-router-dom";
import CloseIcon from '@mui/icons-material/Close';
import Box from "@mui/material/Box";
import EditIcon from '@mui/icons-material/Edit';
import CheckIcon from '@mui/icons-material/Check';
import { Link as RRLink } from "react-router-dom";
import urlJoin from "url-join";
import AddBoxIcon from '@mui/icons-material/AddBox';



export interface MergedDishViewProps {
    mergedDishID : number
}

export function MergedDishView(props :MergedDishViewProps) {

    const authContext = useAuthContext()
    if( authContext === undefined ) {
        console.log("authContext undefined")
    }

    //content of the dish id text field displayed at the bottom
    const [addDishIDTextField,setAddDishIDTextField] = useState<string>("");
    const [mutateError,setMutateError] = useState<ApiError|null>(null)
    const query = useQuery<MergedDishManagementData,ApiError>({
        queryKey: ['getMergedDishes',props.mergedDishID],
        queryFn: () =>
             DefaultService.getMergedDishes(props.mergedDishID),
        onError: (error) => {
            console.log(`getMergedDishes failed with ${error.status} - ${error.message}`)
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
                return
            }
        },
    })


    const queryClient = useQueryClient()

    const updateMergedDish = useMutation<number,ApiError,MergedDishUpdateReq>({
        mutationFn: (req) => {
            return DefaultService.patchMergedDishes(props.mergedDishID, req)
        },
        onSuccess: () => {
            setAddDishIDTextField("")
            //the hard re-fetch in the "then" solves the issue that data would not be re-fetched if there is no cache
            //entry
            //I observed this to happen, when adding to th merged dish in Tab A and viewing the merged dish in tab b
            //Even when refreshing tab b, I did not get a query cache entry in tab B. Maybe this is some weird
            //side effect with react and the query framework?
            return queryClient.invalidateQueries(  {queryKey: ['getMergedDishes',props.mergedDishID]})
                .then(() => query.refetch())
        },
        onError: error => {
            console.log(`createMutation failed with ${error.status} - ${error.message}`)
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
                return
            }
            setMutateError(error)
        }
    })

    const navigate = useNavigate()
    const deleteMergedDishMutation = useMutation<number,ApiError>({
        mutationFn: () => DefaultService.deleteMergedDishes(props.mergedDishID),
        onSuccess: () => {
            queryClient.invalidateQueries(  {queryKey: ['getMergedDishes',props.mergedDishID]})
            navigate("/dishesByDate/today")
        },
        onError: error => {
            console.log(`createMutation failed with ${error.status} - ${error.message}`)
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
                return
            }
            setMutateError(error)
        }
    })


    const [nameEditState, setNameEditState] = useState(false)
    const [editedName,setEditedName] = useState("")

    if( query.isLoading ) {
        return (
            <p>Loading...</p>
        )
    }

    if( query.isError ) {
        return (
            <Alert severity="error"
                   sx={{ mb: 2 }}
            >
                <AlertTitle>{query.error.message }</AlertTitle>
                {query.error.body ? query.error.body.what : "Network error"}
            </Alert>
        )
    }

    const variantLeftColumn = "h6"
    const sxLeftColum : SxProps<Theme> = {
        fontWeight:"medium",
    };

    let nameTableRow = (
        <TableRow>
            <TableCell component={Paper}>
                <Typography variant={variantLeftColumn}>Name</Typography>
            </TableCell>
            <TableCell sx={{display:"flex",justifyContent:"space-between"}} component={Paper}>
                    <Typography sx={sxLeftColum} >{query.data.name}</Typography>
                    <Tooltip title={"Edit"}>
                        <IconButton
                            onClick={() => {
                                setNameEditState(true)
                                setEditedName(query.data.name)
                            }}
                        >
                           <EditIcon/>
                        </IconButton>
                    </Tooltip>
            </TableCell>
        </TableRow>
      
    )
    if( nameEditState ) {
        nameTableRow = (
            <TableRow>
                <TableCell component={Paper}>
                    <Typography variant={variantLeftColumn}>Name</Typography>
                </TableCell>
                <TableCell sx={{ display: "flex", justifyContent: "space-between" }}  component={Paper}>
                        <TextField
                            multiline={true}
                            sx={sxLeftColum}
                            value={editedName}
                            title={"Updated Dish Name"}
                            onChange={(e) => setEditedName(e.target.value)}
                        >

                        </TextField>
                        <Box>
                            <Tooltip title={"Accept"}>
                                <IconButton
                                    onClick={() => {
                                        updateMergedDish.mutate({name: editedName})
                                        setNameEditState(false)
                                    }}
                                >
                                    <CheckIcon />
                                </IconButton>
                            </Tooltip>
                            <Tooltip title={"Cancel"}>
                                <IconButton
                                    onClick={() => {
                                        setNameEditState(false)
                                    }}
                                >
                                    <CloseIcon/>
                                </IconButton>
                            </Tooltip>
                        </Box>
                       
                </TableCell>
            </TableRow>
        )
    }
    return (
        <Container
            maxWidth="sm"
            sx={{
                marginTop:"10px"
            }}>
            { mutateError !== null &&
                <Alert severity="error"
                       action={
                           <IconButton
                               onClick={() => {
                                   setMutateError(null);
                               }}
                           >
                             <CloseIcon></CloseIcon>
                           </IconButton>
                       }
                       sx={{ mb: 2 }}
                >
                <AlertTitle>{mutateError.message }</AlertTitle>
                    {mutateError.body ? mutateError.body.what : "Network error"}
            </Alert>
            }
            <TableContainer component={Paper} elevation={10} >
                <Table>
                    <TableBody>
                        
                        {nameTableRow}

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Served at</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Typography sx={sxLeftColum} >{query.data.servedAt}</Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell   component={Paper}>
                                <Typography  variant={variantLeftColumn}>Contained Dishes</Typography>
                            </TableCell>
                            <TableCell sx={{padding:"0"}} component={Paper}>
                                <List>
                                    {
                                        query.data.containedDishes.sort((a,b) => a.id === b.id ? 0 : a.id < b.id ? -1 : 1  ).map(x => {
                                            return (
                                                <ListItem sx={{ display: "flex", justifyContent: "space-between" }} key={"contained-dish"+x.id}>
                                                    <Tooltip title={"Dish ID"}>
                                                        <ListItemAvatar>
                                                            <Avatar>
                                                                {x.id}
                                                            </Avatar>
                                                        </ListItemAvatar>
                                                    </Tooltip>
                                                    <Link
                                                        component={RRLink}
                                                        to={urlJoin('/dish', x.id.toString())}
                                                    >
                                                        <ListItemText primary={x.name} />
                                                    </Link>
                                                    {query.data.containedDishes.length > 2 &&
                                                        <Tooltip title={"Remove from merged dish"}>
                                                            <IconButton
                                                                aria-label={"remove"}
                                                                onClick={() => updateMergedDish.mutate({addDishIDs:[],removeDishIDs:[x.id]})}
                                                            >
                                                                <DeleteIcon/>
                                                            </IconButton>
                                                        </Tooltip>
                                                    }
                                                </ListItem>
                                            )
                                        })
                                    }
                                </List>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Add Dish</Typography>
                            </TableCell>
                            <TableCell sx={{ display: "flex", justifyContent: "space-between" }}  component={Paper}>
                                <TextField id={"add dish text field"}
                                    variant={"outlined"}
                                    label={"Dish ID"}
                                    inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                                        setAddDishIDTextField(event.target.value);
                                    }}
                                    value={addDishIDTextField}
                                />
                                <Tooltip title={"Add dish"}>
                                    <Box>
                                        <IconButton
                                        disabled={isNaN(+addDishIDTextField) || addDishIDTextField === ""}
                                        onClick={() => updateMergedDish.mutate({ addDishIDs: [Number(addDishIDTextField)], removeDishIDs: [] })}
                                        >
                                        <AddBoxIcon/>
                                    </IconButton>
                                    </Box>
                                    
                                </Tooltip>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell colSpan={2}   >
                                <DeleteButtonWithConfirmation deleteMergedDish={() => deleteMergedDishMutation.mutate()} />
                            </TableCell>
                        </TableRow>

                    </TableBody>
                </Table>
            </TableContainer>
        </Container>
    )
}

interface DeleteButtonWithConfirmationProps {
    deleteMergedDish : () => void
}

function DeleteButtonWithConfirmation(props : DeleteButtonWithConfirmationProps) {
    const [open, setOpen] = useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    return (
        <Box>
            <Button fullWidth={true} variant={"contained"} color={"error"} onClick={handleClickOpen}>
                Delete
            </Button>
            <Dialog
                open={open}
                onClose={handleClose}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    {"Do you really want to delete the merged dish?"}
                </DialogTitle>
                <DialogActions>
                    <Button onClick={() => setOpen(false)}>No</Button>
                    <Button onClick={() => {
                        setOpen(false)
                        props.deleteMergedDish()
                    }} autoFocus>
                        Yes
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
}
