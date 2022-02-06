import axios from "axios";

export type JSONRPCResult<Data> = {
    id: string
    jsonrpc: string
    result: Data
}


// Create a instance of axios to use the same base url.
const axiosAPI = axios.create({
    baseURL: "http://127.0.0.1:5001",
});

export const rpcRequest = <TypeResult = any>(method: string, params: Record<string, any> = {}): Promise<JSONRPCResult<TypeResult>> => {
    const headers = {
        'Content-Type': 'application/json'
    };
    return axiosAPI.post(
        '',
        { "jsonrpc": "2.0", "id": "0", method, params },
        { headers }
    )
        .then(res => {
            return Promise.resolve(res.data);
        })
        .catch(err => {
            return Promise.reject(err);
        });
};