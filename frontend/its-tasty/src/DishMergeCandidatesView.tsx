import {Alert, AlertTitle, Avatar, Checkbox, IconButton, Link, List, ListItem, ListItemAvatar, ListItemText, ListSubheader, Paper, Stack, TextField, Tooltip, Typography} from "@mui/material";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import {
    ApiError,
    CreateMergedDishReq,
    CreateMergedDishResp,
    DefaultService,
    GetDishResp,
    GetMergeCandidatesResp, MergedDishUpdateReq
} from "./services/userAPI";
import {useAuthContext} from "./AuthContext";
import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query";
import {ChangeEvent, useEffect, useState} from "react";
import {SimpleMergedDishView} from "./mergedDishes/simpleMergedDishView";
import {DishVIew} from "./DishView";
import {useNavigate} from "react-router-dom";
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import RestoreIcon from '@mui/icons-material/Restore';
import { Link as RRLink } from "react-router-dom";
import urlJoin from "url-join";


interface CreateButtonProps {
    // the dish on whose page the create button is embedded
    baseDishId: number
    // the other dish required to create a merged dish
    targetDishIds: Set<number>,
    // name for the merged dish
    mergedDishName: string|null
    // Issue data reload in parent component
    refetchCallback : () => Promise<any>

}


export function CreateButton(props: CreateButtonProps) {

    const authContext = useAuthContext()
    const queryClient = useQueryClient();
    if( authContext === undefined ) {
        console.log("authContext undefined")
    }

    const createMutation = useMutation<CreateMergedDishResp,ApiError,string>({
        mutationFn: (mergedDishName ) => {
            const createReq :CreateMergedDishReq = {
                mergedDishes: [...Array.from(props.targetDishIds),props.baseDishId], name: mergedDishName
            };
            return DefaultService.postMergedDishes(createReq)},
        onSuccess: (data) => {

            return Promise.all([
            queryClient.invalidateQueries({queryKey: ['getDishesMergeCandidates', props.baseDishId]}),
            queryClient.invalidateQueries({queryKey: ['getDishes',props.baseDishId]})
            ]).then(() => props.refetchCallback() )
        },
        onError: error => {
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
            }
            console.log(`createMutation failed with ${error.status} - ${error.message}`)
        }
    });


    if (createMutation.isIdle && (props.mergedDishName !== null)) {
        return (
            <Button variant={"contained"} onClick={() => createMutation.mutate(props.mergedDishName!)}>Create New Merged Dish</Button>
        )
    }

    if( createMutation.isLoading ) {
        return (
            <Button variant={"contained"} disabled={true}>"Executing...</Button>
        )
    }

    if( createMutation.isError ) {
        return (
            <Button variant={"contained"} disabled={true}>Error</Button>
        )
    }

    return (
        <Button variant={"contained"} disabled={true}>Create New Merged Dish</Button>
    )

}

interface AddButtonProps {
    //id of the merged dish to which we want to add
    mergedDishID: number
    // id of the dish that we want to add to the merged dish.
    // It is assumed that this is the dish on whose page this button is embedded
    dishID:number
    refetchCallback : () => void

}

function AddButton(props : AddButtonProps) {
    const authContext = useAuthContext()
    const queryClient = useQueryClient();
    if( authContext === undefined ) {
        console.log("authContext undefined")
    }

    interface addMutArgs {
        dishID:number,
        mergedDishID:number
    }
    const addToExistingMutation = useMutation<CreateMergedDishResp,ApiError,addMutArgs>({
        mutationFn: (args) => {
            let req : MergedDishUpdateReq = {
                addDishIDs: [args.dishID]
            }
            return DefaultService.patchMergedDishes(args.mergedDishID,req)
        },
        onSuccess:  () => {
            return Promise.all([
                queryClient.invalidateQueries({queryKey: ['getDishesMergeCandidates', props.dishID]}),
                queryClient.invalidateQueries({queryKey: ['getDishes',props.dishID]}),
                queryClient.invalidateQueries({queryKey: ['getMergedDishes',props.mergedDishID]}),
            ]).then( () => props.refetchCallback() )
        },
        onError: (error) => {
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
            }
        },
    })

    if( addToExistingMutation.isIdle ) {
        return (
            <Button variant={"contained"} onClick={() => addToExistingMutation.mutate( {dishID: props.dishID, mergedDishID:props.mergedDishID} )}>Add to existing</Button>
        )
    }

    if( addToExistingMutation.isLoading ) {
        return (
            <Button variant={"contained"} disabled={true}>"Executing...</Button>
        )
    }

    if( addToExistingMutation.isError) {
        return (
            <Button variant={"contained"} disabled={true}>Error</Button>
        )
    }

    return (
        <Button variant={"contained"} disabled={true}>Add to existing</Button>
    )
}

interface ViewDishMergeCandidatesProps {
    dishID : number
}

export function ViewDishMergeCandidates(props :ViewDishMergeCandidatesProps) {

    const refetchCallback = () => {
        return Promise.all([dishQuery.refetch(), mergeCandidatesQuery.refetch()])
    }
    const authContext = useAuthContext()
    if( authContext === undefined ) {
        console.log("authContext undefined")
    }

    const [nameForMergedDish,setNameForMergedDish] = useState<string|null>(null)
    const navigate = useNavigate()
    const [redirectCountdown,setRedirectCountdown] = useState<number|undefined>(undefined)

    const dishQuery = useQuery<GetDishResp,ApiError,GetDishResp>({
        queryKey: ['getDishes',props.dishID],
        queryFn: () => DefaultService.getDishes(props.dishID),
        onSuccess: (data) => {
            setNameForMergedDish(data.name)
        },
        onError: (error) => {
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
            }
            console.log(`dishQuery failed with ${error.status} - ${error.message}`)
        }
    })

    const mergeCandidatesQuery = useQuery<GetMergeCandidatesResp, ApiError>({
        queryKey: ['getDishesMergeCandidates', props.dishID],
        queryFn: () => DefaultService.getDishesMergeCandidates(props.dishID),
        onError: (error) => {
            if (error.status === 401) {
                authContext?.setAuthStatus(false)
            }
            console.log(`mergeCandidatesQuery failed with ${error.status} - ${error.message}`)
        },
    });


    useEffect(
        () => {
            if( dishQuery.data?.mergedDishID ) {
                if( redirectCountdown === undefined) {
                    setRedirectCountdown(2)
                    return
                }
                if( redirectCountdown === 0 ) {
                    navigate("/mergedDish/"+dishQuery.data.mergedDishID)
                }
                setTimeout(
                    () => {
                        setRedirectCountdown(redirectCountdown-1)
                    },
                    1000
                )
            }
        },
        [dishQuery.data,redirectCountdown,navigate]
    )

    const [selectedDishes,setSelectedDishes] =useState<Set<number>>(new Set<number>() )


    if (dishQuery.isLoading || mergeCandidatesQuery.isLoading ) {
        return (
                <p>Loading... </p>
        )
    }

    if( dishQuery.isError) {
        return (
                <Alert severity="error"
                    sx={{ mb: 2 }}
                >
                <AlertTitle>{dishQuery.error.message}</AlertTitle>
                {dishQuery.error.body ? dishQuery.error.body.what : "Network error"}
                </Alert>
            )
    }


    if( dishQuery.data.mergedDishID && redirectCountdown) {
        return (
            <Container
                maxWidth="sm"
                sx={{
                    marginTop: "10px",
                    display: "flex",
                    flexDirection: "column"
                }}>
                <p>
                    Dish is already part of merged dish. Redirecting to merged dish in {redirectCountdown}s.
                </p>
            </Container>
        )
    }

    if (mergeCandidatesQuery.isError) {
        return (
            <Alert severity="error"
                sx={{ mb: 2 }}
            >
                <AlertTitle>{mergeCandidatesQuery.error.message}</AlertTitle>
                {mergeCandidatesQuery.error.body ? mergeCandidatesQuery.error.body.what : "Network error"}
            </Alert>
        )
    }

    //
    // Process Query Data
    //


    //Split returned ids into dishes that are part of a merged dish and single/individual dishes
    let individualCandidates = mergeCandidatesQuery.data.candidates.filter(x => x.mergedDishID === undefined);
    let uniqueMergedDishCandidates = new Set<number>();
    mergeCandidatesQuery.data.candidates.forEach(x => {
        let mergedDishID = x.mergedDishID;
        if (mergedDishID !== undefined) {
            console.log(`adding mergedDishID: ${x.mergedDishID} from ${JSON.stringify(x)}`)
            uniqueMergedDishCandidates.add(mergedDishID)
        }
    })

    let individualDishesView = (
        <Container
            maxWidth="sm"
            sx={{
                marginTop:"10px",
                display:"flex",
                flexDirection:"column"
            }}>
                <Stack>
                <Typography variant="h5">Create a new merged dish</Typography>
                <p>
                    Select all entries from the list below that you want to go into the
                    new merged dish.
                </p>
                
                <List component={Paper} subheader={<ListSubheader>Candidates</ListSubheader>}>
                    {
                        individualCandidates.map(x =>
                            <ListItem key={x.dishID}>
                                 <Checkbox onChange={(e) => {
                                    setSelectedDishes( prev => {
                                        let updatedSelection = new Set<number>(prev)

                                        if( e.target.checked ) {
                                            updatedSelection.add(x.dishID)
                                        } else {
                                            updatedSelection.delete(x.dishID)
                                            if( nameForMergedDish === x.dishName ) {
                                                setNameForMergedDish(null)
                                            }
                                        }
                                        return updatedSelection
                                    })
                                }} />
                                    <Tooltip title={"Dish ID"}>
                                        <ListItemAvatar>
                                            <Avatar>
                                                {x.dishID}
                                            </Avatar>
                                        </ListItemAvatar>
                                    </Tooltip>
                                    <Link
                                        component={RRLink}
                                        to={urlJoin('/dish', x.dishID.toString())}
                                    >
                                        <ListItemText primary={x.dishName} />
                                    </Link>

                                    <Tooltip title={"Select as name for Merged Dish"}>
                                        <IconButton
                                            onClick={() => setNameForMergedDish(x.dishName)}
                                        >
                                            <ContentCopyIcon/>
                                        </IconButton>
                                    </Tooltip>
                                    </ListItem>
                        )
                    }
                    <Container sx={{ display: "flex", gap:"10px", flexDirection: "row" }}>
                        <TextField id={"newMergedDishNameTextField"}
                            fullWidth={true}
                            variant={"outlined"}
                            label={"Name for Merged Dish"}
                            value={nameForMergedDish === null ? "" : nameForMergedDish}
                            multiline={true}
                            onChange={(event: ChangeEvent<HTMLInputElement>) => {
                                if( event.target.value === "" ) {
                                    setNameForMergedDish(null)
                                } else {
                                    setNameForMergedDish(event.target.value);
                                }
                            }}
                        />
                        <Tooltip title={"Reset merged dish name"}>
                            <IconButton
                                disabled={nameForMergedDish === dishQuery.data.name}
                                onClick={() => {
                                    setNameForMergedDish(dishQuery.data.name)
                                }}
                            >
                                <RestoreIcon/>
                            </IconButton>
                        </Tooltip>
                        <CreateButton
                            mergedDishName={nameForMergedDish}
                            baseDishId={props.dishID}
                            targetDishIds={selectedDishes}
                            refetchCallback={refetchCallback}
                        />
                    </Container>

                </List>
                </Stack>
      
        </Container>

    )


    let mergedDishesView = (
        <Container
            maxWidth="sm"
            sx={{
                marginTop: "10px",
                display: "flex",
                flexDirection: "column"
            }}>
            <Typography variant="h5">Add to existing merged Dish</Typography>
            <p>
                Chose one of the following, already existing merged dishes, that you want to add this dish
                to.
            </p>
            <List  subheader={<ListSubheader>Candidates</ListSubheader>}>
                {
                    Array.from(uniqueMergedDishCandidates).map(mergedDishID =>
                        <ListItem  key={mergedDishID}>
                            <SimpleMergedDishView id={mergedDishID} actionChildren={
                                <AddButton
                                    mergedDishID={mergedDishID}
                                    dishID={props.dishID}
                                    refetchCallback={refetchCallback}
                                />
                            }/>
                           

                        </ListItem>
                    )
                }
            </List>
        </Container>
    )

    return (
        <Stack>
            <DishVIew dishID={props.dishID} showRatingData={false}/>
            {individualCandidates?.length !== 0 && individualDishesView}
            {uniqueMergedDishCandidates.size !== 0 && mergedDishesView}
            { (individualCandidates?.length === 0 && uniqueMergedDishCandidates.size === 0) && 
                <Container
                    maxWidth="sm"
                    sx={{
                        marginTop: "10px",
                        display:"flex",
                        justifyContent:"center"
                    }}>
                    <Typography variant="h6">No Merge Candidates</Typography>
                </Container>
            }
        </Stack>
    );
}