(defcolumns (P :u1) (X :i16) (Y :i16))

(defcall (Y) id (X) (!= 0 (shift P -1)))
