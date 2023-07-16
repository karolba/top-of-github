import { h } from "dom-chef";
import { saveScrollPosition } from "../scrollPosition";

// generates a list of successive positive integers of length n
// range(n) == [0, 1, 2, ..., n-1]
function range(n: number): Array<number> {
    return Array(n)
        .fill(null)
        .map((_, i) => i)
}

// insertEllipsis inserts "..." between array values further than 1 from each other 
// for example, it changes [1, 2, 3, 9, 10, 11, 20, 21]
// into [1, 2, 3, "...", 9, 10, 11, "...", 20, 21]
function insertEllipsis(arr: number[]): (number | string)[] {
    return arr.reduce((result: (number | string)[], current: number, index: number) => {
        if (index === 0 || current - arr[index - 1] === 1) {
            return [...result, current];
        } else {
            return [...result, "...", current];
        }
    }, []);
};

export default function Pagination(page: number, pages: number, onPageChange: (newPage: number) => void): JSX.Element {
    let showNPagesAroundCurrent = 6
    
    // Show more pages around the current one if it's near the start
    if(page < showNPagesAroundCurrent)
        showNPagesAroundCurrent += showNPagesAroundCurrent - page;

    // Show more pages around the current one if it's near the end
    if(pages - page < showNPagesAroundCurrent)
        showNPagesAroundCurrent += showNPagesAroundCurrent - (pages - page);

    const displayPages = range(pages)
        .map(pageNumber => pageNumber + 1) // Page numbers start at 1 to be more human-friendly
        .filter(pageNumber => pageNumber == 1 || pageNumber == pages || Math.abs(pageNumber - page) <= showNPagesAroundCurrent)

    const displayPagesWithEllipsis = insertEllipsis(displayPages)

    return (
        <nav aria-label="Results navigation" className="pt-3 pb-1">
            <ul className="pagination flex-wrap justify-content-center">
                <li className={`page-item ${page == 1 ? 'disabled' : ''}`}>
                    <a className="page-link" href="#" onClick={ev => { ev.preventDefault(); saveScrollPosition(); onPageChange(1) }}>
                        Previous
                    </a>
                </li>
                {displayPagesWithEllipsis
                    .map(pageNumber =>
                        pageNumber == '...'
                        ?
                            <li className="page-item disabled">
                                <a className="page-link" href="#" tabIndex={-1}>...</a>
                            </li>
                        : pageNumber == page
                        ?
                            <li className="page-item active" aria-current="page">
                                <a className="page-link" onClick={()=>false}>{pageNumber}</a>
                            </li>
                        :
                            <li className="page-item">
                                <a className="page-link" href="#" onClick={ev => { ev.preventDefault(); saveScrollPosition(); onPageChange(Number(pageNumber)) }}>
                                    {pageNumber}
                                </a>
                            </li>
                    )
                }
                <li className={`page-item ${page >= pages ? 'disabled' : ''}`}>
                    <a className="page-link" href="#" onClick={ev => { ev.preventDefault(); saveScrollPosition(); onPageChange(pages) }}>
                        Next
                    </a>
                </li>
            </ul>
        </nav>
    )
}