;;error:3:22-23:expected bool, found u1
(defcolumns (P :u1) (X :i16 :padding 1) (Y :i16))
(defcall (Y) dec (X) P)
