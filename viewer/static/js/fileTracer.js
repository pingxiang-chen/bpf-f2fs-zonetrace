const TYPE_UNKNOWN = 0
const TYPE_ROOT = 1
const TYPE_PARENT = 2
const TYPE_FILE = 3
const TYPE_DIRECTORY = 4
const TYPE_HOME = 5

const ICON_UNKNOWN = 'question circle'
const ICON_ROOT = 'disk'
const ICON_PARENT = 'arrow left'
const ICON_FILE = 'file outline'
const ICON_DIRECTORY = 'folder'
const ICON_HOME = 'home'

function getIconType(pathType) {
    return [ICON_UNKNOWN, ICON_ROOT, ICON_PARENT, ICON_FILE, ICON_DIRECTORY, ICON_HOME][pathType]
}

class Blocks {
    constructor(rleCompressedData) {
        this.decompressedData = this.decompressRLE(rleCompressedData); // RLE로 압축된 데이터를 디컴프레스합니다.
    }

    // 주어진 데이터를 디컴프레스하는 함수
    decompressRLE(data) {
        let decompressed = [];
        let byteValue = 0;
        let bitPosition = 7;

        for (let i = 0; i < data.length; i += 2) {
            let bitValue = data[i];
            let count = data[i + 1];

            for (let j = 0; j < count; j++) {
                byteValue |= (bitValue << bitPosition);
                bitPosition--;

                if (bitPosition < 0) {
                    decompressed.push(byteValue);
                    byteValue = 0;
                    bitPosition = 7;
                }
            }
        }

        if (bitPosition < 7) {
            decompressed.push(byteValue);
        }

        return new Uint8Array(decompressed);
    }

    // n번째 인덱스의 비트가 1인지 0인지를 확인하는 메서드
    isBitSet(n) {
        const byteIndex = Math.floor(n / 8);
        const bitIndex = 7 - (n % 8); // 왼쪽부터 0번 비트이므로 7에서 뺍니다.
        return (this.decompressedData[byteIndex] & (1 << bitIndex)) !== 0;
    }
}

document.currentZoneBlocks = null;


document.addEventListener('DOMContentLoaded', function () {


    // 파일 및 폴더 항목을 생성하는 함수
    function createFileSystemItem(item) {
        // 파일 시스템 아이템을 UI에 추가합니다.
        const itemNode = d3.select('#file-system')
            .append('div')
            .attr('class', 'item');

        // 아이콘 유형을 폴더인지 파일인지에 따라 다르게 설정합니다.
        const iconNode = itemNode.append('i')
            .attr('class', `${item.iconType} icon clickable`);

        // 클릭 이벤트를 처리하는 공통 함수
        function handleItemClick(event) {
            // 여기에 클릭 이벤트에 대한 공통 로직을 작성합니다.
            if (item.type === 'folder') {
                const list = d3.select(this.parentNode).select('.list');
                list.style('display', list.style('display') === 'none' ? 'block' : 'none');
            } else if (item.type === TYPE_HOME) {
                updateCurrentFileList(null);
            } else if (item.type !== TYPE_FILE) {
                if (!item.children) {
                    updateCurrentFileList(item);
                }
            } else if (item.type === TYPE_FILE) {
                console.log(item.path);
                getFileInfo(item.path)
            }
        }

        // 클릭 이벤트 리스너를 요소에 추가합니다.
        iconNode.on('click', handleItemClick);

        const content = itemNode.append('div').attr('class', 'content');
        const fileInfo = content.append('div')
            .attr('class', 'file-info clickable')
            .on('click', handleItemClick);  // 같은 핸들러를 사용합니다.

        fileInfo.append('div')
            .attr('class', 'header')
            .text(item.name);

        if (item.size) {
            fileInfo.append('div')
                .attr('class', 'file-size')
                .text(item.size);
        }

        if (item.type === 'folder') {
            // 하위 폴더 및 파일 목록을 생성하고 숨깁니다.
            var list = content.append('div').attr('class', 'list');
            list.selectAll('.item')
                .data(item.children)
                .enter()
                .append(createFileSystemItem);
            list.style('display', 'none');
        }

        return itemNode.node();
    }

    // 파일 시스템을 UI 리스트에 추가하는 함수
    function populateFileSystem(fileSystemData) {
        d3.select('#file-system').selectAll('.item').remove();
        d3.select('#file-system').selectAll('.item')
            .data(fileSystemData)
            .enter()
            .append(createFileSystemItem);
    }

    async function getFileInfo(filePath) {
        const root = await protobuf.load("/static/zns.proto");
        const FileInfoResponse = root.lookupType('FileInfoResponse');
        const response = await fetch(`/api/fileInfo?filePath=${filePath}`);
        const responseData = await response.arrayBuffer();  // Convert response to ArrayBuffer
        const fileInfoResponse = FileInfoResponse.decode(new Uint8Array(responseData));  // Deserialize
        const zoneBitmaps = fileInfoResponse.zoneBitmaps;
        Object.keys(zoneBitmaps).forEach(function (zoneNumber) {
            if (currentZoneId === Number(zoneNumber)) {
                console.log(`zone ${zoneNumber} is set`)
                document.currentZoneBlocks = new Blocks(zoneBitmaps[zoneNumber])
            }
        });
    }

    async function updateCurrentFileList(selectedItem) {
        let nextDirPath = ''
        const isHome = !selectedItem
        if (selectedItem) {
            nextDirPath = selectedItem.path;
        }

        const response = await fetch(`/api/files?dirPath=${nextDirPath}`);
        const data = await response.json()
        const files = data['files'];
        const newFileSystem = [];
        const root = {
            type: TYPE_HOME,
            iconType: ICON_HOME,
            name: '',
            size: '',
            path: '',
            parent: null,
        };
        newFileSystem.push(root);

        if (!isHome && selectedItem && selectedItem.parent) {
            const parent = selectedItem.parent;
            newFileSystem.push({
                type: parent['type'],
                iconType: ICON_PARENT,
                name: '..',
                size: parent['size'],
                path: parent['path'],
                parent: parent.parent || null,
            });
        }

        for (const fileInfo of files) {
            let parent = selectedItem;
            if (!parent) {
                parent = root;
            }
            newFileSystem.push({
                iconType: getIconType(fileInfo['type']),
                type: fileInfo['type'],
                name: fileInfo['name'],
                size: fileInfo['size_str'],
                path: fileInfo['file_path'],
                parent: parent,
            });
        }
        // 파일 시스템 채우기
        populateFileSystem(newFileSystem);
    }


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
        let onChangeBlockSize = () => {
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

        blocksPerLineInput.addEventListener('change', onChangeBlockSize);

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

        const onChangeCellSize = () => {
            cellSize = Number(document.getElementById('cellSizeInput').value);
            reDrawCanvas();
        }
        document.getElementById('cellSizeInput').addEventListener('change', onChangeCellSize);


        // cellColorMap is an array that stores all the blocks within the current zone as a one-dimensional array.
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
                if (document.currentZoneBlocks && document.currentZoneBlocks.isBitSet(colorIndex)) {
                    color = 'green';
                }
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
                document.location.href = `/highlight/${i}`;
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
        updateCurrentFileList(null);

        /* ---------- end of main ---------- */
    }


    main();
});