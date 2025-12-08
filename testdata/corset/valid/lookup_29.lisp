(defpurefun ((i8 :i8 :force) x) x)
(defcolumns (sel :i1) (from :i16) (into :i8))

(defclookup
  l1
  ;; target column
  (into)
  ;; source selector
  sel
  ;; source column
    ((i8 from))
)
