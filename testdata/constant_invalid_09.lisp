;;error:4:18-23:not permitted in pure context
(defcolumns A)
(defun (get) A)
(defconst BROKEN (get))
