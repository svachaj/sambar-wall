IF NOT EXISTS (SELECT 1 FROM sys.tables WHERE name = 't_course_attendance')
BEGIN
    CREATE TABLE t_course_attendance (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        ID_ApplicationForm INT NOT NULL,
        ID_Course INT NOT NULL,
        LessonDate DATE NOT NULL,
        Present BIT NOT NULL,
        UpdatedDate DATETIME NOT NULL,
        CreatedDate DATETIME NOT NULL,
        ID_UpdatedBy INT NOT NULL,
        ID_CreatedBy INT NOT NULL,
        GID UNIQUEIDENTIFIER NOT NULL,
        IsActive BIT NOT NULL CONSTRAINT DF_t_course_attendance_IsActive DEFAULT(1)
    );
END;
GO

IF NOT EXISTS (
    SELECT 1
    FROM sys.indexes
    WHERE name = 'UX_t_course_attendance_App_LessonDate'
      AND object_id = OBJECT_ID('t_course_attendance')
)
BEGIN
    CREATE UNIQUE INDEX UX_t_course_attendance_App_LessonDate
        ON t_course_attendance (ID_ApplicationForm, LessonDate)
        WHERE IsActive = 1;
END;
GO

IF NOT EXISTS (
    SELECT 1
    FROM sys.indexes
    WHERE name = 'IX_t_course_attendance_Course_LessonDate'
      AND object_id = OBJECT_ID('t_course_attendance')
)
BEGIN
    CREATE INDEX IX_t_course_attendance_Course_LessonDate
        ON t_course_attendance (ID_Course, LessonDate);
END;
GO
