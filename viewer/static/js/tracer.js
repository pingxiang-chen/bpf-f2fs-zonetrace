document.addEventListener('DOMContentLoaded', function () {
    console.log('DOM loaded with JavaScript');

    /**
     * Get the last segment from the specified URL path.
     *
     * This function takes a URL as an argument, splits it into segments,
     * and returns the last non-empty segment.
     *
     * @param {string} url - The URL to be parsed.
     * @returns {string} - The last non-empty segment of the URL.
     */
    function getLastPathSegment(url) {
        const segments = url.split('/');
        // Return the last segment. If the last character is '/',
        // it will return an empty string, so return the second last segment instead
        return segments.pop() || segments.pop();
    }

    let cellSize = 1;
    let blocksPerLine = 0;
    let zoomLevel = 1;

    const currentZoneId = Number(getLastPathSegment(document.location.pathname)) // Retrieves the number corresponding to :num in the URL `/zone/:num`
    let lastSegmentType = -2; // -2: NotChanged
    const curSegTypeSpan = document.getElementById('curSegType') // Span element to display the current zone's segment type

    /**
     * Get information for a specific zone.
     * This function returns information about the number of zones, segments in each zone, etc.
     *
     * @param {string} zoneId - The number of the zone to get information for
     * @returns {Promise<object>} - A Promise object containing an object with zone information
     * @throws {Error} - Throws an error if getting zone information fails
     */
    async function getZoneInfo(zoneId) {
        const response = await fetch(`/api/info/${zoneId}`);
        if (!response.ok) {
            throw new Error('Cannot get zone info');
        }
        return await response.json();
    }

    /**
     * Get the label for a given segment type.
     *
     * @param {number} segmentType - The segment type (-2 to 5, inclusive)
     * @returns {string} - The label for the segment type
     */
    function getSegmentTypeLabel(segmentType) {
        // -2: NotChanged, -1: Unknown, 0: HotData, 1: WarmData, 2: ColdData, 3: HotNode, 4: WarmNode, 5: ColdNode, 6: Empty
        return ['NotChanged', 'Unknown', 'Hot Data', 'Warm Data', 'Cold Data', 'Hot Node', 'Warm Node', 'Cold Node', 'Empty'][segmentType + 2];
    }

    /**
     * Get the color for a given segment type.
     *
     * @param {number} segmentType - The segment type (-2 to 5, inclusive)
     * @returns {string} - The color for the segment type
     */
    function getSegmentTypeColor(segmentType) {
        if (segmentType < 0 || 6 < segmentType) return 'black';
        return ['red', 'yellow', 'blue', 'pink', 'orange', 'skyblue', 'black'][segmentType];
    }

    /**
     * Update the current zone's segment type and display it on the screen.
     *
     * @param {number} segmentType - The segment type to update
     */
    async function updateCurrentSegmentType(segmentType) {
        if (lastSegmentType === segmentType || segmentType === -2) return;
        lastSegmentType = segmentType;
        curSegTypeSpan.innerText = getSegmentTypeLabel(segmentType);
    }

    function getDivisors(num) {
        const divisors = [];
        for (let i = 1; i <= Math.sqrt(num); i++) {
            if (num % i === 0) {
                divisors.push(i);
                if (num / i !== i) divisors.push(num / i);
            }
        }
        // divisors.sort((a, b) => a - b);
        return divisors;
    }

    const onChangeCellSize = () => {
        cellSize = Number(document.getElementById('cellSizeInput').value);
        reDrawCanvas();
    }

    let onChangeBlockSize = () => {
        // it will be re-declared later
    }

    let reDrawCanvas = async () => {
        // it will be re-declared later
    };

    /**
     * Main function.
     * Performs major logic and initializes the page and data display.
     */
    async function main() {
        const info = await getZoneInfo(currentZoneId);
        document.getElementById('curZoneNum').innerText = String(currentZoneId);
        updateCurrentSegmentType(info.last_segment_type);
        const bitmapSize = info.block_per_segment;
        const maxSegmentNumber = info.total_segment_per_zone;

        // set blocksPerLineInput
        const blocksPerLineInput = document.getElementById('blocksPerLineInput');
        blocksPerLine = bitmapSize
        blocksPerLineInput.value = bitmapSize;
        blocksPerLineInput.setAttribute("max", bitmapSize);
        const blockDivisors = getDivisors(bitmapSize);
        onChangeBlockSize = () => {
            const currentValue = Number(blocksPerLineInput.value);
            const isIncreased = blocksPerLine < currentValue;
            let next;
            if (isIncreased) {
                next = (blockDivisors.filter(d => d >= currentValue).sort((a, b) => a - b))[0];
            } else {
                next = (blockDivisors.filter(d => d <= currentValue).sort((a, b) => b - a))[0];
            }
            blocksPerLine = next;
            blocksPerLineInput.value = blocksPerLine;
            zoomLevel = bitmapSize / blocksPerLine;
            reDrawCanvas();
        }

        /* ---------- Draw the canvas representing zone's blocks ---------- */

        // Canvas size
        let canvasRowSize = bitmapSize / zoomLevel;
        let canvasColSize = maxSegmentNumber * zoomLevel;

        // Create the canvas
        let canvas = d3.select("#chartCanvas")
            .attr("width", canvasRowSize * cellSize)
            .attr("height", canvasColSize * cellSize)
            .node();
        let context = canvas.getContext("2d");

        reDrawCanvas = async () => {
            context.clearRect(0, 0, canvas.width, canvas.height);
            canvasRowSize = bitmapSize / zoomLevel;
            canvasColSize = maxSegmentNumber * zoomLevel;
            canvas = d3.select("#chartCanvas")
                .attr("width", canvasRowSize * cellSize)
                .attr("height", canvasColSize * cellSize)
                .node();
            context = canvas.getContext("2d");
            cellColorMap.forEach((color, i) => {
                const [xPos, yPos] = getDrawPos(Math.floor(i / bitmapSize), i % bitmapSize, zoomLevel)
                context.fillStyle = color;
                context.fillRect(xPos, yPos, cellSize, cellSize);
            })
        }

        const cellColorMap = Array.from({length: maxSegmentNumber * bitmapSize}, () => "white");

        function getColorMapIndex(segmentNumber, bitmapIndex) {
            return segmentNumber * bitmapSize + bitmapIndex;
        }

        function getDrawPos(segmentNumber, bitmapIndex) {
            let index1D = segmentNumber * bitmapSize + bitmapIndex;
            let newRowSize = bitmapSize / zoomLevel;
            let yPos = Math.floor(index1D / newRowSize);
            let xPos = index1D % newRowSize;
            return [xPos * cellSize, yPos * cellSize];
        }

        /**
         * Draw a segment's information on the canvas.
         *
         * @param {number[]} row - Bitmap array of the segment
         * @param {number} y - Y coordinate position on the canvas where the segment is drawn
         * @param {number} segmentType - The segment's type represented by a number
         */
        function drawChart(row, y, segmentType) {
            row.forEach((d, i) => {
                let color = "white";
                if (d === 1) {
                    color = getSegmentTypeColor(segmentType);
                } else if (d === -1) {
                    color = "white";
                }
                const colorIndex = getColorMapIndex(y, i);
                if (cellColorMap[colorIndex] === color) {
                    return;
                }
                cellColorMap[colorIndex] = color;
                const [xPos, yPos] = getDrawPos(y, i)
                context.fillStyle = color;
                context.fillRect(xPos, yPos, cellSize, cellSize);
            });
        }


        /* ---------- Draw SVG ---------- */

        // Close the connection when clicking on another zone using AbortController, signal
        const ctx = new AbortController();
        const signal = ctx.signal;

        const margin = {top: 30, right: 25 + 50, bottom: 30, left: 40}
        const width = 450 + 50 - margin.left - margin.right
        const height = 800 - margin.top - margin.bottom

        const svg = d3.select("#zones")
            .append("svg")
            .attr("width", width + margin.left + margin.right)
            .attr("height", height + margin.top + margin.bottom)
            .append("g")
            .attr("transform", `translate(${margin.left}, ${margin.top})`);

        const zoneTotalSize = info.total_zone;
        let xLength = 15; // Display 15 zones in a column by default
        let yLength = Math.ceil(zoneTotalSize / xLength);
        xLength = Math.ceil(zoneTotalSize / yLength);

        if (xLength > yLength) {
            // If xLength is greater than yLength, swap them to draw in a tall format
            [xLength, yLength] = [yLength, xLength]
        }

        const xVars = Array.from({length: xLength}, (_, i) => i);
        const yVars = Array.from({length: yLength}, (_, i) => i * xLength);
        yVars.sort((a, b) => a - b)

        // Build X scales and axis:
        const xScale = d3.scaleBand()
            .domain(xVars)
            .range([0, width])
            .padding(0.05);
        svg.append("g")
            .style("font-size", 15)
            .attr("transform", `translate(0, ${height})`)
            .call(d3.axisBottom(xScale).tickSize(0))
            .select(".domain").remove()

        // Build Y scales and axis:
        const yScale = d3.scaleBand()
            .domain(yVars)
            .range([0, height])
            .padding(0.05);
        svg.append("g")
            .style("font-size", 15)
            .call(d3.axisLeft(yScale).tickSize(0))
            .select(".domain").remove()

        // create a tooltip
        const tooltip = d3.select("#zones")
            .append("div")
            .style("opacity", 0)
            .attr("class", "tooltip")
            .style("position", "absolute")
            .style("z-index", "10")
            .style("background-color", "white")
            .style("border", "solid")
            .style("border-width", "2px")
            .style("border-radius", "5px")
            .style("padding", "5px")


        // Three function that change the tooltip when user hover / move / leave a cell
        const mouseover = function (event, d) {
            if (zoneTotalSize <= d || d === currentZoneId) {
                return
            }
            tooltip
                .style("opacity", 1)
            d3.select(this)
                .style("stroke", "black")
                .style("opacity", 1)
        }
        const mousemove = function (event, d) {
            tooltip
                .html("zone: " + d)
                .style("top", (event.y + 5) + "px")
                .style("left", (event.x + 30) + "px")
        }
        const mouseleave = function (event, d) {
            if (zoneTotalSize <= d || d === currentZoneId) {
                return
            }
            tooltip
                .style("opacity", 0)
            d3.select(this)
                .style("stroke", "none")
                .style("opacity", 0.8)
        }

        const zoneCell = Array.from({length: xLength * yLength}, (_, i) => i);

        // add the squares
        svg.selectAll()
            .data(zoneCell, function (v, i) {
                return i;
            })
            .join("rect")
            .attr("x", function (v, i) {
                return xScale(xVars[Math.floor(i % (zoneCell.length / yVars.length))])
            })
            .attr("y", function (v, i) {
                return yScale(yVars[Math.floor(i / (zoneCell.length / yVars.length))])
            })
            .attr("width", xScale.bandwidth())
            .attr("height", yScale.bandwidth())
            .style("fill", function (v, i) {
                if (zoneTotalSize <= i) {
                    return "#f3f3f3"
                }
                return "black";
            })
            .style("stroke-width", 4)
            .style("stroke", function (v, i) {
                if (i === currentZoneId) {
                    return "black"
                }
                return "none";
            })
            .style("opacity", 0.8)
            .on("mouseover", mouseover)
            .on("mousemove", mousemove)
            .on("mouseleave", mouseleave)
            .on("click", function (event, i) {
                if (i === currentZoneId) {
                    return
                }
                ctx.abort()
                document.location.href = `/zone/${i}`;
            })


        // Add title to graph
        svg.append("text")
            .attr("x", 0)
            .attr("y", -10)
            .attr("text-anchor", "left")
            .style("font-size", "22px")
            .text("Zones");

        const lastUpdateZone = {};

        /**
         * Updates the color or text when the segmentType of the corresponding zone changes.
         *
         * @param {number} zoneNo - The zone number.
         * @param {number} segmentType - The segmentType that has been updated.
         */
        async function updateZoneSegmentType(zoneNo, segmentType) {
            if (segmentType === -2) {
                return
            }
            if (zoneNo === currentZoneId) {
                updateCurrentSegmentType(segmentType)
            }
            let color = getSegmentTypeColor(segmentType);
            const _now = Date.now();
            lastUpdateZone[zoneNo] = _now
            const cell = svg.selectAll("rect")
                .filter(function (v, i) {
                    return i === zoneNo;
                })
            cell.style("fill", color)

            // Change it back to black after 1 second when it has been updated
            setTimeout(() => {
                if (lastUpdateZone[zoneNo] === _now) {
                    cell.style("fill", "black")
                }
            }, 1000)
        }


        /**
         * Connect to the server and receive zone updates in real-time,
         * triggering various events accordingly.
         */
        async function handleStreamData() {
            // Decode data from the server as protobuf.
            const root = await protobuf.load("/static/zns.proto");
            const ZoneResponse = root.lookupType('ZoneResponse');

            const response = await fetch(`/api/zone/${currentZoneId}`, {signal});
            const reader = response.body.getReader();

            let buf = [];
            while (true) {
                const {done, value} = await reader.read();
                if (done) {
                    break;
                }
                buf = buf.concat(Array.from(value));
                while (true) {
                    const r = protobuf.Reader.create(buf)
                    let zone;
                    try {
                        zone = ZoneResponse.decodeDelimited(r);
                    } catch (e) {
                        break;
                    }
                    buf = buf.slice(r.pos);

                    // Perform zone updates with data decoded using protobuf
                    if (zone.lastSegmentType !== -2) { // -2: NotChanged
                        updateZoneSegmentType(zone.zoneNo, zone.lastSegmentType);
                    }
                    if (!zone.segments || zone.segments.length === 0) {
                        continue;
                    }
                    // Draw with the bitmap data of each segment
                    zone.segments.forEach((segment, _) => {
                        let segmentBitmap = [];
                        if (segment.map) {
                            for (let b of Object.values(segment.map)) {
                                for (let i = 0; i < 8; i++) {
                                    segmentBitmap.push((b >> i) & 1);
                                }
                            }
                        } else {
                            segmentBitmap = Array.from({length: info.block_per_segment}, () => -1);
                        }
                        const ping = Date.now() - zone.time;
                        // console.log(`received ${segment.segmentNo}, ping: ${ping}ms`)
                        drawChart(segmentBitmap, segment.segmentNo, segment.segmentType);
                    })
                }
            }
            // Stream ended
        }

        handleStreamData();

        /* ---------- end of main ---------- */
    }


    main();
});