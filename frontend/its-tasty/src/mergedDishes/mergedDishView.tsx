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
    Theme, TextField, AlertTitle, Alert
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

    const updateContainedDishes = useMutation<number,ApiError,MergedDishUpdateReq>({
        mutationFn: (req) => {
            return DefaultService.patchMergedDishes(props.mergedDishID, req)
        },
        onSuccess: () => {
            //the hard re-fetch in the then solves the issue that data would not be re-fetched if there is no cache
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
                        <TableRow>
                            <TableCell  component={Paper}>
                                <Typography variant={variantLeftColumn}>Name</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Typography sx={sxLeftColum} >{query.data.name}</Typography>
                            </TableCell>
                        </TableRow>

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
                            <TableCell component={Paper}>
                                <List>
                                    {
                                        query.data.containedDishes.sort((a,b) => a.id === b.id ? 0 : a.id < b.id ? -1 : 1  ).map(x => {
                                            return (
                                                <ListItem key={"contained-dish"+x.id}>
                                                    <Tooltip title={"Dish ID"}>
                                                        <ListItemAvatar>
                                                            <Avatar>
                                                                {x.id}
                                                            </Avatar>
                                                        </ListItemAvatar>
                                                    </Tooltip>
                                                    <ListItemText primary={x.name} />
                                                    {query.data.containedDishes.length > 2 &&
                                                        <Tooltip title={"Remove from merged dish"}>
                                                            <IconButton
                                                                aria-label={"remove"}
                                                                onClick={() => updateContainedDishes.mutate({addDishIDs:[],removeDishIDs:[x.id]})}
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
                    </TableBody>
                </Table>
            </TableContainer>
            <Container sx={{mt:"5px",display:"flex",justifyContent:"space-evenly",flex:"flex-grow"}}>
                <TextField id={"add dish text field"}
                           variant={"outlined"}
                           label={"Dish ID"}
                           inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
                           onChange={(event: ChangeEvent<HTMLInputElement>) => {
                               setAddDishIDTextField(event.target.value);
                           }}
                />
                <Button
                    variant={"contained"}
                    disabled={ isNaN(+addDishIDTextField) || addDishIDTextField === "" }
                    onClick={() => updateContainedDishes.mutate({addDishIDs:[Number(addDishIDTextField)],removeDishIDs:[]})}

                >
                    Add Dish
                </Button>

            </Container>
            <Container sx={{mt:"5px",display:"flex",justifyContent:"center",flex:"flex-grow"}}>
                <DeleteButtonWithConfirmation deleteMergedDish={() => deleteMergedDishMutation.mutate() }/>
            </Container>
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
            <Button variant={"contained"} color={"error"} onClick={handleClickOpen}>
                Delete Merged Dish
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
