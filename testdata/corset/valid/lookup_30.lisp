(defpurefun ((i8 :i8 :force) x) x)
(module m1)
(defcolumns (sel :i1) (X :i16) (Y :i8))

(defclookup
  l1
  ;; target column
  (m2.X m2.Y)
  ;; source selector
  m1.sel
  ;; source column
  ((i8 m1.X) m1.Y))

(module m2)
(defcolumns (X :i8) (Y :i8))
