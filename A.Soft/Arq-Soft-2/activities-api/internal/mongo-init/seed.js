db = db.getSiblingDB('sporthub');
db.activities.insertOne({title: "Clase Demo", description: "Actividad inicial", ownerUserId: "1"});
