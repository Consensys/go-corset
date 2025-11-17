(defcolumns (P :u1) (X :i16 :padding 1) (Y :i16))
(defcall (Y) dec (X) (!= 0 P))
