const TYPE_UNKNOWN = 0
const TYPE_ROOT = 1
const TYPE_PARENT = 2
const TYPE_FILE = 3
const TYPE_DIRECTORY = 4


const ICON_UNKNOWN = 'question circle'
const ICON_ROOT = 'disk'
const ICON_PARENT = 'arrow left'
const ICON_FILE = 'file'
const ICON_DIRECTORY = 'folder'

document.addEventListener('DOMContentLoaded', function () {

    function getIconType(pathType) {
        return [ICON_UNKNOWN, ICON_ROOT, ICON_PARENT, ICON_FILE, ICON_DIRECTORY][pathType]
    }

    // 예시 파일 시스템 데이터
    const fileSystem = [
        {iconType: 'arrow left', name: '..'},
        {
            iconType: 'folder',
            name: 'Documents',
            children: [
                {iconType: 'file', name: 'report.pdf', size: '200KB'},
                {iconType: 'file', name: 'essay.docx', size: '1MB'}
            ]
        },
        {
            iconType: 'folder',
            name: 'Music',
            children: [
                {iconType: 'file', name: 'song.mp3', size: '5MB'}
            ]
        },
        {iconType: 'file', name: 'todo.txt', size: '50KB'}
    ];

    // 파일 및 폴더 항목을 생성하는 함수
    function createFileSystemItem(item) {
        // 파일 시스템 아이템을 UI에 추가합니다.
        const itemNode = d3.select('#file-system')
            .append('div')
            .attr('class', 'item');

        // 아이콘 유형을 폴더인지 파일인지에 따라 다르게 설정합니다.
        itemNode.append('i')
            .attr('class', `${item.iconType} icon clickable`);

        const content = itemNode.append('div').attr('class', 'content');
        const fileInfo = content.append('div')
            .attr('class', 'file-info clickable')
            .on('click', function (event) {
                if (item.type === 'folder') {
                    // 폴더를 클릭할 때 하위 목록의 표시 여부를 전환합니다.
                    const list = d3.select(this.parentNode).select('.list');
                    list.style('display', list.style('display') === 'none' ? 'block' : 'none');
                } else if (item.type !== TYPE_FILE) {
                    if (!item.children) {
                        updateCurrentFileList(item);
                    }
                } else if (item.type === TYPE_FILE) {
                    console.log(item.path)
                }

            });

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

    async function updateCurrentFileList(newPathItem) {
        console.log('newPathItem', newPathItem)
        let nextDirPath = ''
        if (newPathItem) {
            nextDirPath = newPathItem.path;
        }
        const response = await fetch(`/api/files?dirPath=${nextDirPath}`);
        const data = await response.json()
        const files = data['files'];
        if (newPathItem) {
            const prevDir = newPathItem
            fileSystem.length = 0; // clear fileSystem
            fileSystem.push({
                type: prevDir['type'],
                iconType: 'arrow left',
                name: prevDir['name'],
                size: prevDir['size_str'],
                path: prevDir['file_path'],
            });
        } else {
            fileSystem.length = 0; // clear fileSystem
        }

        for (const fileInfo of files) {
            fileSystem.push({
                iconType: getIconType(fileInfo['type']),
                type: fileInfo['type'],
                name: fileInfo['name'],
                size: fileInfo['size_str'],
                path: fileInfo['file_path'],
            });
        }
        // 파일 시스템 채우기
        populateFileSystem(fileSystem);
    }

    updateCurrentFileList('');

});