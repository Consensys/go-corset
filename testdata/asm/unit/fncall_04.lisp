(defcolumns (P :u1) (X :i16 :padding 1) (Y :i16))

(defun (selector-with-prev)
    (* (- 1 (shift P -1)) P))

(defcall (Y) id (X) (!= 0 (selector-with-prev)))
