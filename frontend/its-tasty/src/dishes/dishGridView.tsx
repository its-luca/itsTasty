import {
    DataGrid,
    GridColDef,
    GridRenderCellParams,
    GridRowsProp, GridValidRowModel
} from '@mui/x-data-grid';
import type {} from '@mui/x-data-grid/themeAugmentation';
import {
    useMutation,
    useQuery,
    useQueryClient,
} from "@tanstack/react-query";
import {
    ApiError,
    CreateMergedDishReq,
    CreateMergedDishResp,
    DefaultService,
    GetAllDishesResponse,
} from "../services/userAPI";
import {ChangeEvent, useState} from "react";
import {useAuthContext} from "../AuthContext";
import IconButton from "@mui/material/IconButton";
import CloseIcon from "@mui/icons-material/Close";
import {Alert, AlertTitle, Link, Stack, TextField} from "@mui/material";
import Container from "@mui/material/Container";
import Button from "@mui/material/Button";
import urlJoin from "url-join";
import { Link as RRLink} from "react-router-dom";
import Tooltip from "@mui/material/Tooltip";
import SearchIcon from '@mui/icons-material/Search';


const columns: GridColDef[] = [
    { field: 'colID', headerName: 'ID', width: 150,
        renderCell: (params : GridRenderCellParams) => (
            <Link
                component={RRLink}
                to={urlJoin('/dish',params.value.toString())}
                sx={{fontSize:"large",fontWeight:"bold",textAlign:"center"}} >
                {params.value}
            </Link>
        )
    },
    {
        field: 'colName', headerName: 'Name', flex:1,
    },
    { field: 'colServedAt', headerName: 'Location', width: 150 },
    { field: 'colMergedDishID', headerName: 'Belongs to merged dish?', flex:1,
        renderCell: (params : GridRenderCellParams) => {
            if(params.value ) {
                return <Link
                    component={RRLink}
                    to={urlJoin('/mergedDish',params.value.toString())}
                    sx={{fontSize:"large",fontWeight:"bold",textAlign:"center"}} >
                    {params.value}
                </Link>
            }

            return (
                <Container >
                    No
                    <Tooltip title={"Search Merge Candidates"}>
                        <IconButton
                            component={RRLink}
                            to={urlJoin('/mergeCandidates',params.id.toString())}>
                            <SearchIcon/>
                        </IconButton>
                    </Tooltip>

                </Container>

            )
        }
    },


];

export function DishGridView() {

    const authContext = useAuthContext()
    if( authContext === undefined ) {
        console.log("authContext undefined")
    }

    const queryClient = useQueryClient();
    const allDishIDs = useQuery<GetAllDishesResponse,ApiError>({
        queryKey: ['getGetAllDishes'],
        queryFn: () => DefaultService.getGetAllDishes(),
        onError: (error) => {
            console.log(`getMergedDishes failed with ${error.status} - ${error.message}`)
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
                return
            }
        },
    })

    const [mutateError,setMutateError] = useState<ApiError|null>(null);
    const createMutation = useMutation<CreateMergedDishResp,ApiError,CreateMergedDishReq>({
        mutationFn: ( req) => {
            return DefaultService.postMergedDishes(req)},
        onSuccess: (data) => {
            setMutateError(null)
            return Promise.all([
                queryClient.invalidateQueries({queryKey: ['getGetAllDishes']}).then(() => allDishIDs.refetch())
            ])
        },
        onError: error => {
            if( error.status === 401) {
                authContext?.setAuthStatus(false)
            }
            console.log(`createMutation failed with ${error.status} - ${error.message}`)
            setMutateError(error)
        }
    });



    const [selectedRows,setSelectedRows] = useState<GridValidRowModel[]>([]);
    const [nameForMergedDish,setNamForMergedDish] = useState<string>("");


    if( allDishIDs.isLoading ) {
        return (
            <p>Loading...</p>
        )
    }

    if( allDishIDs.error) {
        return (
            <Alert severity="error"
                   sx={{ mb: 2 }}
            >
                <AlertTitle>{allDishIDs.error.message }</AlertTitle>
                {allDishIDs.error.body ? allDishIDs.error.body.what : "Network error"}
            </Alert>
        )
    }

    const mutateErrorView =
        <Alert severity="error"
               sx={{ mb: 2 }}
               action={
                    <IconButton onClick={() => {
                            setMutateError(null);
                        }}
                    >
                        <CloseIcon></CloseIcon>
                    </IconButton>
                }>
        <AlertTitle>{createMutation.error?.message ? createMutation.error.message : "Unknown Error" }</AlertTitle>
            {createMutation.error?.body.what}
    </Alert>



    const rows : GridRowsProp = allDishIDs.data.data.map( simpleDishEntry => {
        return {
            id: simpleDishEntry.id,
            colID: simpleDishEntry.id,
            colName: simpleDishEntry.name,
            colServedAt: simpleDishEntry.servedAt,
            colMergedDishID: simpleDishEntry.mergedDishID
        }
    })


    return (
        <Stack sx={{ maxHeight:'100%'}}>
            {mutateError && mutateErrorView}
            <Container sx={{margin:"10px", display:"flex",justifyContent:"space-around"}}
            >
                <TextField id={"dishGridViewMergedDishNameTextField"}
                           fullWidth={true}
                           variant={"outlined"}
                           label={"Name for merged Dish"}
                           disabled={selectedRows.length < 1}
                           value={nameForMergedDish}
                           onChange={(event: ChangeEvent<HTMLInputElement>) => {
                               event.stopPropagation()
                               setNamForMergedDish(event.target.value);
                           }}
                />
                    <Button
                        variant={"outlined"}
                        disabled={ nameForMergedDish === "" || selectedRows.length < 2}
                        onClick={() => {
                            const req : CreateMergedDishReq = {
                                name: nameForMergedDish,
                                mergedDishes: selectedRows.map( row => row.colID)
                            }
                            createMutation.mutate(req)
                        }}
                    >
                        Merge Dishes from selected Rows
                    </Button>

            </Container>


            <DataGrid
                rows={rows}
                columns={columns}
                autoHeight={true}
                checkboxSelection
                disableRowSelectionOnClick
                onRowSelectionModelChange={ (model,details) => {
                    if( selectedRows.length > 0 && nameForMergedDish !== "" && model.length === 0 ) {
                        setNamForMergedDish("")
                    }
                    const selectedIds = new Set(model)
                    const selection = rows.filter(row => selectedIds.has(row.id))
                    if( selectedRows.length === 0 && nameForMergedDish === "" && model.length > 0 ) {
                        setNamForMergedDish(selection[0].colName)
                    }
                    setSelectedRows(selection)
                }}
                initialState={{
                    pagination: {paginationModel:{pageSize: 25}}
                }}
            />
        </Stack>
    )
}
