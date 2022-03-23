const sample = [
  { x: 772, y: 341, delay: 693 },
  { x: 773, y: 341, delay: 15 },
  { x: 774, y: 341, delay: 5 },
  { x: 775, y: 341, delay: 8 },
  { x: 776, y: 341, delay: 18 },
  { x: 777, y: 341, delay: 12 },
  { x: 778, y: 341, delay: 10 },
  { x: 778, y: 342, delay: 2 },
  { x: 779, y: 342, delay: 6 },
  { x: 780, y: 342, delay: 16 },
  { x: 781, y: 342, delay: 4 },
  { x: 782, y: 342, delay: 6 },
  { x: 783, y: 342, delay: 4 },
  { x: 784, y: 342, delay: 9 },
  { x: 785, y: 342, delay: 3 },
  { x: 786, y: 342, delay: 2 },
  { x: 787, y: 342, delay: 4 },
  { x: 788, y: 342, delay: 4 },
  { x: 789, y: 342, delay: 2 },
  { x: 789, y: 341, delay: 2 },
  { x: 790, y: 341, delay: 2 },
  { x: 791, y: 341, delay: 2 },
  { x: 793, y: 341, delay: 7 },
  { x: 794, y: 341, delay: 2 },
  { x: 795, y: 341, delay: 2 },
  { x: 795, y: 340, delay: 2 },
  { x: 796, y: 340, delay: 4 },
  { x: 797, y: 340, delay: 2 },
  { x: 798, y: 340, delay: 2 },
  { x: 799, y: 340, delay: 3 },
  { x: 800, y: 340, delay: 3 },
  { x: 801, y: 340, delay: 2 },
  { x: 803, y: 340, delay: 4 },
  { x: 805, y: 340, delay: 6 },
  { x: 806, y: 340, delay: 2 },
  { x: 807, y: 340, delay: 2 },
  { x: 808, y: 340, delay: 4 },
  { x: 809, y: 340, delay: 2 },
  { x: 810, y: 340, delay: 2 },
  { x: 811, y: 340, delay: 2 },
  { x: 813, y: 340, delay: 4 },
  { x: 814, y: 340, delay: 2 },
  { x: 815, y: 340, delay: 2 },
  { x: 816, y: 340, delay: 4 },
  { x: 817, y: 340, delay: 2 },
  { x: 819, y: 340, delay: 2 },
  { x: 820, y: 340, delay: 4 },
  { x: 821, y: 339, delay: 2 },
  { x: 822, y: 339, delay: 2 },
  { x: 823, y: 339, delay: 2 },
  { x: 824, y: 339, delay: 4 },
  { x: 825, y: 339, delay: 2 },
  { x: 826, y: 339, delay: 2 },
  { x: 827, y: 339, delay: 2 },
  { x: 828, y: 339, delay: 4 },
  { x: 829, y: 339, delay: 2 },
  { x: 830, y: 339, delay: 2 },
  { x: 831, y: 339, delay: 2 },
  { x: 832, y: 339, delay: 4 },
  { x: 833, y: 339, delay: 2 },
  { x: 834, y: 339, delay: 2 },
  { x: 835, y: 339, delay: 2 },
  { x: 836, y: 339, delay: 4 },
  { x: 838, y: 339, delay: 2 },
  { x: 839, y: 339, delay: 4 },
  { x: 840, y: 339, delay: 2 },
  { x: 841, y: 339, delay: 2 },
  { x: 842, y: 339, delay: 2 },
  { x: 843, y: 339, delay: 2 },
  { x: 844, y: 339, delay: 4 },
  { x: 845, y: 339, delay: 2 },
  { x: 846, y: 339, delay: 2 },
  { x: 847, y: 339, delay: 2 },
  { x: 848, y: 339, delay: 4 },
  { x: 849, y: 339, delay: 2 },
  { x: 850, y: 339, delay: 2 },
  { x: 851, y: 339, delay: 2 },
  { x: 852, y: 339, delay: 4 },
  { x: 853, y: 339, delay: 2 },
  { x: 854, y: 339, delay: 2 },
  { x: 855, y: 339, delay: 2 },
  { x: 856, y: 339, delay: 4 },
  { x: 857, y: 339, delay: 3 },
  { x: 858, y: 339, delay: 1 },
  { x: 859, y: 339, delay: 2 },
  { x: 860, y: 339, delay: 6 },
  { x: 861, y: 339, delay: 2 },
  { x: 862, y: 339, delay: 2 },
  { x: 863, y: 339, delay: 5 },
  { x: 864, y: 339, delay: 6 },
  { x: 865, y: 339, delay: 4 },
  { x: 866, y: 339, delay: 4 },
  { x: 867, y: 339, delay: 4 },
  { x: 868, y: 339, delay: 8 },
  { x: 869, y: 339, delay: 4 },
  { x: 870, y: 339, delay: 2 },
  { x: 871, y: 339, delay: 4 },
  { x: 872, y: 339, delay: 10 },
  { x: 873, y: 339, delay: 2 },
  { x: 874, y: 339, delay: 4 },
  { x: 875, y: 339, delay: 4 },
  { x: 876, y: 339, delay: 8 },
  { x: 877, y: 339, delay: 4 },
  { x: 878, y: 339, delay: 4 },
  { x: 879, y: 339, delay: 4 },
  { x: 880, y: 339, delay: 8 },
  { x: 881, y: 339, delay: 4 },
  { x: 882, y: 339, delay: 2 },
  { x: 883, y: 339, delay: 4 },
  { x: 884, y: 339, delay: 6 },
  { x: 885, y: 339, delay: 5 },
  { x: 886, y: 339, delay: 3 },
  { x: 887, y: 339, delay: 2 },
  { x: 888, y: 339, delay: 10 },
  { x: 889, y: 339, delay: 5 },
  { x: 890, y: 339, delay: 5 },
  { x: 891, y: 339, delay: 6 },
  { x: 892, y: 339, delay: 8 },
  { x: 893, y: 339, delay: 6 },
  { x: 894, y: 339, delay: 4 },
  { x: 895, y: 339, delay: 6 },
  { x: 896, y: 339, delay: 10 },
  { x: 897, y: 339, delay: 7 },
  { x: 898, y: 339, delay: 8 },
  { x: 899, y: 339, delay: 10 },
  { x: 900, y: 339, delay: 24 },
  { x: 901, y: 339, delay: 12 },
  { x: 902, y: 339, delay: 14 },
  { x: 903, y: 339, delay: 12 },
  { x: 904, y: 339, delay: 22 },
  { x: 905, y: 339, delay: 8 },
  { x: 906, y: 339, delay: 10 },
  { x: 907, y: 339, delay: 14 },
  { x: 908, y: 339, delay: 25 },
  { x: 909, y: 339, delay: 14 },
  { x: 910, y: 339, delay: 10 },
  { x: 911, y: 339, delay: 12 },
  { x: 911, y: 338, delay: 10 },
  { x: 912, y: 338, delay: 14 },
  { x: 913, y: 338, delay: 14 },
  { x: 914, y: 338, delay: 18 },
  { x: 915, y: 338, delay: 18 },
  { x: 916, y: 338, delay: 34 },
  { x: 917, y: 338, delay: 24 },
  { x: 918, y: 338, delay: 45 },
  { x: 919, y: 338, delay: 46 },
  { x: 920, y: 338, delay: 69 },
  { x: 921, y: 338, delay: 31 },
  { x: 922, y: 338, delay: 103 },
  { x: 923, y: 338, delay: 181 },
  { x: 923, y: 339, delay: 95 },
  { x: 924, y: 339, delay: 220 },
];

const genScales = () => {
  var scales = [];
  for (var i = 0; i < sample.length - 1; i++) {
    scales.push(
      (sample[i + 1].x - sample[i].x) /
        (sample[sample.length - 1].x - sample[0].x)
    );
  }
  return scales;
};

export const genMouseScale = (x1, x2) => {
  var points = [];
  var lastX = x1;
  const scales = genScales();
  const diff = parseFloat((x2 - x1).toFixed(2));
  scales.forEach((scale, i) => {
    const x = lastX + diff * scale;
    points.push({
      x: x,
      y: sample[i].y,
      delay: sample[i].delay,
    });
    lastX = x;
  });
  return points;
};

const graphPointsScale = (events) => {
  const startX = events[0].x;
  events.forEach((e) => {
    let xrep = "";
    for (var i = 0; i < Math.floor(e.x) - startX; i++) {
      xrep += " ";
    }
    console.log(xrep + "|");
    sleep(e.delay);
  });
};

function sleep(milliseconds) {
  const date = Date.now();
  let currentDate = null;
  do {
    currentDate = Date.now();
  } while (currentDate - date < milliseconds);
}

// graphPoints(genPoints(300, 400));
// module.exports = { genMouseScale };
