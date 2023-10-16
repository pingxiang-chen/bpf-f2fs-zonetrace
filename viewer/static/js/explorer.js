const TYPE_UNKNOWN = 0
const TYPE_ROOT = 1
const TYPE_PARENT = 2
const TYPE_FILE = 3
const TYPE_DIRECTORY = 4
const TYPE_HOME = 5

const ICON_UNKNOWN = 'question circle'
const ICON_ROOT = 'disk'
const ICON_PARENT = 'arrow left'
const ICON_FILE = 'file'
const ICON_DIRECTORY = 'folder'
const ICON_HOME = 'home'

document.addEventListener('DOMContentLoaded', function () {

    function getIconType(pathType) {
        return [ICON_UNKNOWN, ICON_ROOT, ICON_PARENT, ICON_FILE, ICON_DIRECTORY, ICON_HOME][pathType]
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
        const iconNode = itemNode.append('i')
            .attr('class', `${item.iconType} icon clickable`);

        // 클릭 이벤트 핸들러를 아이콘에도 추가합니다.
        iconNode.on('click', function (event) {
            // 이 부분에 원하는 동작을 구현합니다.
            // 예를 들어, 폴더 아이콘을 클릭했을 때의 동작을 여기에 추가할 수 있습니다.
            // 아래는 'file-info' 요소에 추가된 이벤트 핸들러의 로직을 반복하는 예입니다.
            if (item.type === 'folder') {
                const list = d3.select(this.parentNode).select('.list');
                list.style('display', list.style('display') === 'none' ? 'block' : 'none');
            } else if (item.type === TYPE_HOME) {
                console.log('home');
                updateCurrentFileList(null);
            } else if (item.type !== TYPE_FILE) {
                if (!item.children) {
                    updateCurrentFileList(item);
                }
            } else if (item.type === TYPE_FILE) {
                console.log(item.path);
            }
        });

        const content = itemNode.append('div').attr('class', 'content');
        const fileInfo = content.append('div').attr('class', 'file-info clickable')

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

    async function updateCurrentFileList(selectedItem) {
        let nextDirPath = ''
        let prevDir;
        const isHome = !selectedItem
        if (selectedItem) {
            nextDirPath = selectedItem.path;
            prevDir = selectedItem.parent;
        }

        const response = await fetch(`/api/files?dirPath=${nextDirPath}`);
        const data = await response.json()
        const files = data['files'];
        const newFileSystem = [];
        newFileSystem.push({
            type: TYPE_HOME,
            iconType: ICON_HOME,
            name: '',
            size: '',
            path: '',
            parent: null,
        });

        if (!isHome && prevDir) {
            newFileSystem.push({
                type: prevDir['type'],
                iconType: ICON_PARENT,
                name: '..',
                size: prevDir['size'],
                path: prevDir['path'],
            });
        }

        for (const fileInfo of files) {
            newFileSystem.push({
                iconType: getIconType(fileInfo['type']),
                type: fileInfo['type'],
                name: fileInfo['name'],
                size: fileInfo['size_str'],
                path: fileInfo['file_path'],
                parent: selectedItem,
            });
        }
        // 파일 시스템 채우기
        populateFileSystem(newFileSystem);
    }

    updateCurrentFileList(null);

});