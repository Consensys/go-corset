(defcolumns (P :u1) (X :i16) (Y :i16))

(defcall (Y) id ((shift X -1)) (!= 0 P))
