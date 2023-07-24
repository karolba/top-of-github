export function humanizeNumber(num: number): string {
    const suffixes = ["", " thousand", " million", " billion", " trillion"];
    const numStr = num.toString();
    const numDigits = numStr.length;

    if (numDigits <= 4) {
        return numStr;
    } else {
        const suffixIndex = Math.floor((numDigits - 1) / 3);
        const normalizedNum = num / Math.pow(10, suffixIndex * 3);
        const significantDigits = normalizedNum.toPrecision(3); // Show up to 3 significant digits

        return significantDigits.toString() + suffixes[suffixIndex];
    }
}