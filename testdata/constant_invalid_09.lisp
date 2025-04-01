;;error:4:19-22:not permitted in pure context
(defcolumns (A :i16))
(defun (get) A)
(defconst BROKEN (get))
