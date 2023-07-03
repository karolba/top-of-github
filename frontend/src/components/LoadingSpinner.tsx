import { h } from "dom-chef"


export default function LoadingSpinner(): JSX.Element {

    return (
        <div className="text-center">
            <div className="spinner-grow m-5 p-3" role="status">
                <span className="visually-hidden">Loading...</span>
            </div>
        </div>
    )
}
