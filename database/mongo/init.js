// Inicialización de MongoDB para activities-api y bookings-api

// Conectar a la base de datos
db = db.getSiblingDB('gym_db');

// Crear colección de activities con validación de esquema
db.createCollection('activities', {
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['owner_id', 'name', 'category', 'duration', 'price', 'status'],
            properties: {
                owner_id: {
                    bsonType: ['int', 'long'],
                    description: 'ID del usuario admin que creó la actividad'
                },
                name: {
                    bsonType: 'string',
                    description: 'Nombre de la actividad'
                },
                description: {
                    bsonType: 'string'
                },
                category: {
                    bsonType: 'string',
                    enum: ['Yoga', 'Spinning', 'Pilates', 'Funcional', 'CrossFit', 'Zumba', 'Boxing', 'Natacion'],
                    description: 'Categoría de la actividad'
                },
                duration: {
                    bsonType: 'int',
                    minimum: 30,
                    maximum: 180,
                    description: 'Duración en minutos'
                },
                price: {
                    bsonType: 'double',
                    minimum: 0,
                    description: 'Precio de la actividad'
                },
                image_url: {
                    bsonType: 'string'
                },
                status: {
                    bsonType: 'string',
                    enum: ['active', 'inactive'],
                    description: 'Estado de la actividad'
                },
                created_at: {
                    bsonType: 'date'
                },
                updated_at: {
                    bsonType: 'date'
                }
            }
        }
    }
});

// Crear índices para activities
db.activities.createIndex({ "owner_id": 1 });
db.activities.createIndex({ "name": 1 });
db.activities.createIndex({ "category": 1 });
db.activities.createIndex({ "status": 1 });

// Crear colección de schedules con validación
db.createCollection('schedules', {
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['activity_id', 'instructor', 'day_of_week', 'start_time', 'end_time', 'location', 'max_capacity', 'status'],
            properties: {
                activity_id: {
                    bsonType: 'string',
                    description: 'ID de la actividad asociada'
                },
                instructor: {
                    bsonType: 'string',
                    description: 'Nombre del instructor'
                },
                day_of_week: {
                    bsonType: 'string',
                    enum: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'],
                    description: 'Día de la semana'
                },
                start_time: {
                    bsonType: 'string',
                    pattern: '^([0-1][0-9]|2[0-3]):[0-5][0-9]$',
                    description: 'Hora de inicio (formato HH:MM)'
                },
                end_time: {
                    bsonType: 'string',
                    pattern: '^([0-1][0-9]|2[0-3]):[0-5][0-9]$',
                    description: 'Hora de fin (formato HH:MM)'
                },
                location: {
                    bsonType: 'string',
                    description: 'Ubicación de la clase'
                },
                max_capacity: {
                    bsonType: 'int',
                    minimum: 1,
                    description: 'Capacidad máxima'
                },
                current_bookings: {
                    bsonType: 'int',
                    minimum: 0,
                    description: 'Reservas actuales'
                },
                status: {
                    bsonType: 'string',
                    enum: ['active', 'cancelled', 'full'],
                    description: 'Estado del horario'
                },
                created_at: {
                    bsonType: 'date'
                },
                updated_at: {
                    bsonType: 'date'
                }
            }
        }
    }
});

// Crear índices para schedules
db.schedules.createIndex({ "activity_id": 1 });
db.schedules.createIndex({ "day_of_week": 1, "start_time": 1 });
db.schedules.createIndex({ "location": 1, "day_of_week": 1, "start_time": 1 }, { unique: true });
db.schedules.createIndex({ "instructor": 1 });
db.schedules.createIndex({ "status": 1 });

// Crear colección de bookings con validación
db.createCollection('bookings', {
    validator: {
        $jsonSchema: {
            bsonType: 'object',
            required: ['user_id', 'schedule_id', 'status'],
            properties: {
                user_id: {
                    bsonType: ['int','long'],
                    description: 'ID del usuario que reserva'
                },
                schedule_id: {
                    bsonType: 'string',
                    description: 'ID del horario reservado'
                },
                activity_name: {
                    bsonType: 'string'
                },
                activity_category: {
                    bsonType: 'string'
                },
                instructor: {
                    bsonType: 'string'
                },
                day_of_week: {
                    bsonType: 'string'
                },
                start_time: {
                    bsonType: 'string'
                },
                end_time: {
                    bsonType: 'string'
                },
                location: {
                    bsonType: 'string'
                },
                price: {
                    bsonType: 'double'
                },
                status: {
                    bsonType: 'string',
                    enum: ['confirmed', 'cancelled', 'attended'],
                    description: 'Estado de la reserva'
                },
                created_at: {
                    bsonType: 'date'
                },
                updated_at: {
                    bsonType: 'date'
                }
            }
        }
    }
});

// Crear índices para bookings

// NOTA: El índice único de user_id + schedule_id se crea en init-indexes.js como índice parcial
// para permitir re-inscripciones después de cancelar
db.bookings.createIndex({ "user_id": 1 });
db.bookings.createIndex({ "schedule_id": 1 });
db.bookings.createIndex({ "status": 1 });
db.bookings.createIndex({ "user_id": 1, "schedule_id": 1 }, { unique: true }); // Usuario no puede reservar 2 veces el mismo horario

print('MongoDB inicializado correctamente para gym-booking-system');