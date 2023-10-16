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
    // const fileSystem = [
    //     {iconType: 'arrow left', name: '..'},
    //     {
    //         iconType: 'folder',
    //         name: 'Documents',
    //         children: [
    //             {iconType: 'file', name: 'report.pdf', size: '200KB'},
    //             {iconType: 'file', name: 'essay.docx', size: '1MB'}
    //         ]
    //     },
    //     {
    //         iconType: 'folder',
    //         name: 'Music',
    //         children: [
    //             {iconType: 'file', name: 'song.mp3', size: '5MB'}
    //         ]
    //     },
    //     {iconType: 'file', name: 'todo.txt', size: '50KB'}
    // ];

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

    updateCurrentFileList(null);
});