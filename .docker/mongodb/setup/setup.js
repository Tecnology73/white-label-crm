(async function() {
    rs.initiate({
        _id: 'rs0',
        members: [
            {
                _id: 0,
                host: 'localhost:27017',
                priority: 1,
            },
        ],
    });

    while (true) {
        const { ismaster } = db.isMaster();

        if (ismaster) break;

        await new Promise(resolve => setTimeout(resolve, 1e3));
    }

    db.getSiblingDB('admin').createUser({ user: 'root', pwd: 'root', roles: [{ role: 'root', db: 'admin' }]});
})()
