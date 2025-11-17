(module sq)
(defcolumns (X :i32) (Y :i32))
;; Y = (X + X) % 2^32
(defcall (Y) add (X X))
