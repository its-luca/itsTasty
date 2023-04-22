import {useAuthContext} from "../AuthContext";
import {useQuery} from "@tanstack/react-query";
import {ApiError, DefaultService, MergedDishManagementData} from "../services/userAPI";
import {
    Avatar,
    Container,
    List, ListItem,
    ListItemAvatar,
    ListItemText,
    Paper,
    Skeleton,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableRow,
    Tooltip
} from "@mui/material";
import ErrorOutlineIcon from "@mui/icons-material/ErrorOutline";
import Typography from "@mui/material/Typography";

interface SimpleMergedDishViewProps {
    id : number
    actionChildren : React.ReactElement | null
}


export function SimpleMergedDishView(props : SimpleMergedDishViewProps) {

    const authContext = useAuthContext();

    const query = useQuery<MergedDishManagementData,ApiError>({
        queryKey:['getMergedDishes',props.id],
        queryFn: () => {
            return DefaultService.getMergedDishes(props.id)},
        onError: (e) => {
            if( e.status === 401) {
                authContext?.setAuthStatus(false)
            }
        }
    })


    if( query.isLoading ) {
        return (
                <Container  component={Paper} elevation={10} sx={{mt:"10px",display:"flex",justifyContent:"space-between"}} >
                    <Skeleton variant={"text"} height={"60px"} width={"60px"} sx={{padding:"20px"}}/>
                    <Skeleton variant={"text"} height={"60px"} width={"120px"}/>
                </Container>
        )
    }

    if( query.error ) {
        return (
            <Container  component={Paper} elevation={10} sx={{mt:"10px",display:"flex",justifyContent:"center"}} >
                <ErrorOutlineIcon fontSize={"large"}/>
            </Container>
        )
    }


    return (
    <TableContainer component={Paper} elevation={10} >
        <Table>
            <TableBody>
                <TableRow >
                    <TableCell  align="center" colSpan={2} >
                        <Typography sx={{fontWeight:"medium"}} >{query.data.name}</Typography>
                    </TableCell>
                </TableRow>

                <TableRow>
                    <TableCell >
                        <Typography variant={"h6"}>Contained Dishes</Typography>
                    </TableCell>
                    <TableCell >
                        <List>
                            {
                                    query.data.containedDishes.sort((a, b) => a.id === b.id ? 0 : a.id < b.id ? -1 : 1).map(x =>
                                    <ListItem key={"containedDishIDs"+x.id}>
                                                    <Tooltip title={"Dish ID"}>
                                                        <ListItemAvatar>
                                                            <Avatar>
                                                                {x.id}
                                                            </Avatar>
                                                        </ListItemAvatar>
                                                    </Tooltip>
                                                    <ListItemText primary={x.name} />
                                                   
                                                </ListItem>
                                )
                            }
                        </List>
                    </TableCell>
                </TableRow>
                {props.actionChildren && 
                 <TableRow >
                        <TableCell align="right" colSpan={2}>
                            {props.actionChildren}
                    </TableCell>
                </TableRow>
                }
               
            </TableBody>
        </Table>
    </TableContainer>
    )
}