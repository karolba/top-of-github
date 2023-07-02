import { h } from "dom-chef";

// generates a list of successive positive integers of length n
// range(n) == [0, 1, 2, ..., n-1]
function range(n: number): Array<number> {
    return Array(n)
        .fill(null)
        .map((_, i) => i)
}

export default function Pagination(page: number, pages: number, onPageChange: (newPage: number) => void): JSX.Element {
    return (
        <nav aria-label="Results navigation">
            <ul className="pagination flex-wrap">
                <li className={`page-item ${page == 1 ? 'disabled' : ''}`}>
                    <a className="page-link" href="#" onClick={(ev)=>{ ev.preventDefault(); onPageChange(1) }}>Previous</a>
                </li>
                {range(pages)
                    .map(pageNumber => pageNumber + 1) // Page numbers start at 1 to be more human-friendly
                    .map(pageNumber =>
                        pageNumber == page
                        ?
                            <li className="page-item active" aria-current="page">
                                <a className="page-link" onClick={()=>false}>{pageNumber}</a>
                            </li>
                        :
                            <li className="page-item">
                                <a className="page-link" href="#" onClick={(ev)=>{ ev.preventDefault(); onPageChange(pageNumber) }}>{pageNumber}</a>
                            </li>
                    )
                }
                <li className={`page-item ${page >= pages ? 'disabled' : ''}`}>
                    <a className="page-link" href="#" onClick={(ev)=>{ ev.preventDefault(); onPageChange(pages) }}>Next</a>
                </li>
            </ul>
        </nav>
    )
}