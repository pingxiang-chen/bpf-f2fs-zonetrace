document.addEventListener('DOMContentLoaded', function () {
    let currentPath = "";

    function getIconType(pathType) {
        switch (pathType) {
            case 0: // UnknownPath
                return 'question circle';
            case 1: // RootPath
                return 'home';
            case 2: // ParentPath
                return 'arrow left';
            case 3: // FilePath
                return 'file';
            case 4: // DirectoryPath
                return 'folder';
        }
        return 'question circle'; // UnknownPath
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
                    var list = d3.select(this.parentNode).select('.list');
                    list.style('display', list.style('display') === 'none' ? 'block' : 'none');
                }
                if (!item.children) {
                    currentPath = item.path;
                    updateCurrentFileList(currentPath);
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

    async function updateCurrentFileList(dirPath) {
        const response = await fetch(`/api/files?dirPath=${dirPath}`);
        const data = await response.json()
        fileSystem.length = 0; // clear fileSystem
        console.log(data)
        for (const fileInfo of data['files']) {
            fileSystem.push({
                iconType: getIconType(fileInfo['type']),
                name: fileInfo['name'],
                size: fileInfo['size_str'],
                path: fileInfo['file_path'],
            });
        }
        // 파일 시스템 채우기
        populateFileSystem(fileSystem);
    }

    updateCurrentFileList(currentPath);

});