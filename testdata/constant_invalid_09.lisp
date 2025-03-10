;;error:4:18-23:not permitted in pure context
(defcolumns (A :i16))
(defun (get) A)
(defconst BROKEN (get))
