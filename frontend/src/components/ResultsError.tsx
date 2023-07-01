import { h } from "dom-chef"


export default function ResultsError(error: any): JSX.Element {
    return (
        <div className="text-center">
            <h3>Error</h3>
            <code>{error}</code>
        </div>
    )
}
