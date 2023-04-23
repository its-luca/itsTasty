import {
    DataGrid,
    GridColDef,
    GridPaginationModel,
    GridRenderCellParams,
    GridRowSelectionModel,
    GridRowsProp
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
        onSuccess:(data) => {
            const tmp: GridRowsProp = data.data.map(simpleDishEntry => {
                return {
                    id: simpleDishEntry.id,
                    colID: simpleDishEntry.id,
                    colName: simpleDishEntry.name,
                    colServedAt: simpleDishEntry.servedAt,
                    colMergedDishID: simpleDishEntry.mergedDishID
                }
            })
            setGridRows(tmp)

        },
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

            //update local data without having to refetch from server
            const selectedIds = new Set(gridRowSelectionModel)
            setGridRows( (prevRows) => {
                return prevRows.map( (entry,index) => {
                    return selectedIds.has(index) ? {...entry, colMergedDishID: data.mergedDishID} : entry
                })
            })

            /*
            //https://mui.com/x/react-data-grid/row-updates/#the-updaterows-method says this should work but apparently you need premium after all
            gridRowSelectionModel.forEach( rowID => {
                apiRef.current.updateRows([{ id: rowID, colMergedDishID: data.mergedDishID}])

            })*/

            updateGridRowSelectionModel([])
            return Promise.all([
                queryClient.invalidateQueries({queryKey: ['getGetAllDishes']})
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




    const [nameForMergedDish,setNameForMergedDish] = useState<string>("");

    const [gridPaginationModel,setGridPaginationModel] = useState<GridPaginationModel>({pageSize:25,page:0})
    const [gridRowSelectionModel, setGridRowSelectionModel] = useState<GridRowSelectionModel>([])
    const updateGridRowSelectionModel = (updatedModel : GridRowSelectionModel) => {
        if (updatedModel.length === 0) {
            setNameForMergedDish("")
        }
        const selectedIds = new Set(updatedModel)
        const selection = gridRows.filter(row => selectedIds.has(row.id))
        if (gridRowSelectionModel.length === 0 && nameForMergedDish === "" && updatedModel.length > 0) {
            setNameForMergedDish(selection[0].colName)
        }
        setGridRowSelectionModel(updatedModel)
    }
    const [gridRows, setGridRows] = useState < GridRowsProp>([]);



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
            {createMutation.error?.body ? createMutation.error.body.what : "Network error"}
    </Alert>


 

    return (
        <Stack sx={{ maxHeight:'100%'}}>
            {mutateError && mutateErrorView}
            <Container sx={{margin:"10px",gap:"10px", display:"flex",justifyContent:"space-around",alignItems:"center"}}
            >
                <TextField id={"dishGridViewMergedDishNameTextField"}
                           fullWidth={true}
                           variant={"outlined"}
                           label={"Name for merged Dish"}
                           disabled={gridRowSelectionModel.length < 1}
                           value={nameForMergedDish}
                           onChange={(event: ChangeEvent<HTMLInputElement>) => {
                               event.stopPropagation()
                               setNameForMergedDish(event.target.value);
                           }}
                />
                    <Button
                        variant={"outlined"}
                    disabled={nameForMergedDish === "" || gridRowSelectionModel.length < 2}
                        onClick={() => {
                            const selectedIds = new Set(gridRowSelectionModel)
                            const req : CreateMergedDishReq = {
                                name: nameForMergedDish,
                                mergedDishes: gridRows.filter((_,index) => selectedIds.has(index)).map(entry => entry.colID)
                            }
                            createMutation.mutate(req)
                        }}
                    >
                    {"Merge " + gridRowSelectionModel.length + " Dishes from selected Rows"}
                    </Button>
                    <Button
                        variant={"outlined"}
                        disabled={gridRowSelectionModel.length === 0}
                        color={"error"}
                        onClick={() => updateGridRowSelectionModel([]) }
                    >
                        Clear Selection
                    </Button>

            </Container>


            <DataGrid
                rows={gridRows}
                columns={columns}
                autoHeight={true}
                checkboxSelection
                disableRowSelectionOnClick
                rowSelectionModel={gridRowSelectionModel}
                onRowSelectionModelChange={ updatedModel => updateGridRowSelectionModel(updatedModel) }
                paginationModel={gridPaginationModel}
                onPaginationModelChange={model => {
                    setGridPaginationModel(model)
                }}

                initialState={{
                    sorting: {
                        sortModel: [
                            {
                                field: "colID",
                                sort: "desc",
                            }
                        ]
                    }
                }}
            />
        </Stack>
    )
}
